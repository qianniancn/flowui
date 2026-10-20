package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// tokenCounts is one record's token use. CacheWrite1h is the part of
// CacheWrite cached for an hour and billed at a different price.
type tokenCounts struct {
	input, output, cacheRead, cacheWrite, cacheWrite1h int64
}

func (t tokenCounts) total() int64 {
	return t.input + t.output + t.cacheRead + t.cacheWrite
}

func (t tokenCounts) plus(other tokenCounts) tokenCounts {
	return tokenCounts{
		input:        t.input + other.input,
		output:       t.output + other.output,
		cacheRead:    t.cacheRead + other.cacheRead,
		cacheWrite:   t.cacheWrite + other.cacheWrite,
		cacheWrite1h: t.cacheWrite1h + other.cacheWrite1h,
	}
}

// modelPrice is a model's list price in dollars per million tokens.
type modelPrice struct {
	input, output, cacheRead, cacheWrite, cacheWrite1h float64
}

func (p modelPrice) cost(t tokenCounts) float64 {
	cacheWrite5m := t.cacheWrite - t.cacheWrite1h
	return (float64(t.input)*p.input +
		float64(t.output)*p.output +
		float64(t.cacheRead)*p.cacheRead +
		float64(cacheWrite5m)*p.cacheWrite +
		float64(t.cacheWrite1h)*p.cacheWrite1h) / 1e6
}

var prices = map[string]modelPrice{
	"claude-sonnet-4-5": {input: 3, output: 15, cacheRead: .30, cacheWrite: 3.75, cacheWrite1h: 6},
	"claude-opus-4-1":   {input: 15, output: 75, cacheRead: 1.5, cacheWrite: 18.75, cacheWrite1h: 30},
	"gpt-5-codex":       {input: 1.25, output: 10, cacheRead: .125},
	"gpt-5":             {input: 1.25, output: 10, cacheRead: .125},
	"deepseek-chat":     {input: .27, output: 1.10, cacheRead: .07},
	"deepseek-reasoner": {input: .55, output: 2.19, cacheRead: .14},
}

const pricesAsOf = "2026-10-20"

// record is one priced model call. The demo generates these from session-file
// shaped data; the aggregation below does not know where they came from.
type record struct {
	at       time.Time
	provider string
	session  string
	model    string
	tokens   tokenCounts
}

type providerUse struct {
	provider string
	sessions int
	cost     float64
	tokens   tokenCounts
	daily    []float64
}

type modelUse struct {
	provider string
	model    string
	priced   bool
	cost     float64
	tokens   tokenCounts
}

type dayUse struct {
	cost   float64
	tokens tokenCounts
}

// report is the same shape as pitwall's flow.Report: totals, a provider
// split with per-day costs, a model split, and one entry per day.
type report struct {
	days      []time.Time
	cost      float64
	sessions  int
	tokens    tokenCounts
	unpriced  int64
	savings   float64
	providers []providerUse
	models    []modelUse
	daily     []dayUse
}

func newReport(records []record, now time.Time, days int) report {
	location := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	since := today.AddDate(0, 0, -days+1)
	result := report{
		days:  make([]time.Time, days),
		daily: make([]dayUse, days),
	}
	for i := range days {
		result.days[i] = since.AddDate(0, 0, i)
	}

	type sessionKey struct{ provider, session string }
	providerSeen := map[string]map[sessionKey]bool{}
	providers := map[string]*providerUse{}
	models := map[string]*modelUse{}

	for _, rec := range records {
		stamp := rec.at.In(location)
		day := time.Date(stamp.Year(), stamp.Month(), stamp.Day(), 0, 0, 0, 0, location)
		index := int(math.Round(day.Sub(since).Hours() / 24))
		if index < 0 || index >= days {
			continue
		}

		price, priced := prices[rec.model]
		callCost := 0.0
		if priced {
			callCost = price.cost(rec.tokens)
			result.savings += float64(rec.tokens.cacheRead) *
				(price.input - price.cacheRead) / 1e6
		} else {
			result.unpriced += rec.tokens.total()
		}

		result.cost += callCost
		result.tokens = result.tokens.plus(rec.tokens)
		result.daily[index].cost += callCost
		result.daily[index].tokens = result.daily[index].tokens.plus(rec.tokens)

		provider := providers[rec.provider]
		if provider == nil {
			provider = &providerUse{provider: rec.provider, daily: make([]float64, days)}
			providers[rec.provider] = provider
		}
		if providerSeen[rec.provider] == nil {
			providerSeen[rec.provider] = map[sessionKey]bool{}
		}
		key := sessionKey{rec.provider, rec.session}
		if !providerSeen[rec.provider][key] {
			providerSeen[rec.provider][key] = true
			provider.sessions++
			result.sessions++
		}
		provider.cost += callCost
		provider.tokens = provider.tokens.plus(rec.tokens)
		provider.daily[index] += callCost

		model := models[rec.model]
		if model == nil {
			model = &modelUse{provider: rec.provider, model: rec.model, priced: priced}
			models[rec.model] = model
		}
		model.cost += callCost
		model.tokens = model.tokens.plus(rec.tokens)
	}

	for _, provider := range providers {
		result.providers = append(result.providers, *provider)
	}
	sort.Slice(result.providers, func(i, j int) bool {
		return providerLess(result.providers[i], result.providers[j])
	})
	for _, model := range models {
		result.models = append(result.models, *model)
	}
	sort.Slice(result.models, func(i, j int) bool {
		return modelLess(result.models[i], result.models[j])
	})
	return result
}

func providerLess(a, b providerUse) bool {
	if a.cost != b.cost {
		return a.cost > b.cost
	}
	if a.tokens.total() != b.tokens.total() {
		return a.tokens.total() > b.tokens.total()
	}
	return a.provider < b.provider
}

func modelLess(a, b modelUse) bool {
	if a.cost != b.cost {
		return a.cost > b.cost
	}
	if a.tokens.total() != b.tokens.total() {
		return a.tokens.total() > b.tokens.total()
	}
	return a.model < b.model
}

func providerName(key string) string {
	switch key {
	case "claude":
		return "Claude Code"
	case "codex":
		return "Codex"
	case "deepseek":
		return "DeepSeek"
	default:
		return key
	}
}

// demoRecords builds deterministic session-shaped data for 90 days. The
// example reads no files: these records stand in for what pitwall's scanner
// would return from Claude Code, Codex, and pi sessions.
func demoRecords(now time.Time, days int) []record {
	rng := rand.New(rand.NewSource(20261020))
	location := now.Location()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	start := end.AddDate(0, 0, -days+1)
	profiles := []struct {
		provider string
		models   []string
		sessions int
	}{
		{"claude", []string{"claude-sonnet-4-5", "claude-sonnet-4-5", "claude-opus-4-1"}, 2},
		{"codex", []string{"gpt-5-codex", "gpt-5-codex", "gpt-5"}, 1},
		{"deepseek", []string{"deepseek-chat", "deepseek-chat", "deepseek-reasoner"}, 2},
	}

	records := make([]record, 0, days*10)
	for offset := range days {
		day := start.AddDate(0, 0, offset)
		for _, profile := range profiles {
			sessions := profile.sessions + rng.Intn(2)
			for session := range sessions {
				id := fmt.Sprintf("%s-%s-%d", profile.provider, day.Format("20060102"), session)
				model := profile.models[rng.Intn(len(profile.models))]
				calls := 1 + rng.Intn(2)
				for range calls {
					at := day.Add(time.Duration(8+rng.Intn(11))*time.Hour +
						time.Duration(rng.Intn(60))*time.Minute)
					records = append(records, record{
						at:       at,
						provider: profile.provider,
						session:  id,
						model:    model,
						tokens:   demoTokens(rng, model),
					})
				}
			}
		}
	}
	return records
}

func demoTokens(rng *rand.Rand, model string) tokenCounts {
	switch {
	case strings.HasPrefix(model, "claude-opus"):
		input := 150_000 + rng.Int63n(500_000)
		output := 10_000 + rng.Int63n(35_000)
		cacheRead := 300_000 + rng.Int63n(1_300_000)
		cacheWrite := 30_000 + rng.Int63n(100_000)
		return tokenCounts{
			input: input, output: output, cacheRead: cacheRead, cacheWrite: cacheWrite,
			cacheWrite1h: cacheWrite * int64(25+rng.Intn(50)) / 100,
		}
	case strings.HasPrefix(model, "claude"):
		input := 350_000 + rng.Int63n(1_100_000)
		output := 20_000 + rng.Int63n(70_000)
		cacheRead := 1_000_000 + rng.Int63n(4_500_000)
		cacheWrite := 50_000 + rng.Int63n(220_000)
		return tokenCounts{
			input: input, output: output, cacheRead: cacheRead, cacheWrite: cacheWrite,
			cacheWrite1h: cacheWrite * int64(25+rng.Intn(50)) / 100,
		}
	case strings.HasPrefix(model, "deepseek"):
		return tokenCounts{
			input:     1_500_000 + rng.Int63n(6_000_000),
			output:    80_000 + rng.Int63n(260_000),
			cacheRead: 2_000_000 + rng.Int63n(10_000_000),
		}
	default:
		return tokenCounts{
			input:     350_000 + rng.Int63n(1_100_000),
			output:    15_000 + rng.Int63n(55_000),
			cacheRead: 500_000 + rng.Int63n(2_500_000),
		}
	}
}

func money(value float64) string {
	formatted := fmt.Sprintf("%.2f", value)
	return "$" + thousands(formatted[:len(formatted)-3]) + formatted[len(formatted)-3:]
}

func axisMoney(value float64) string {
	if value == 0 {
		return "$0"
	}
	if value >= 10 {
		return "$" + thousands(fmt.Sprintf("%.0f", value))
	}
	return money(value)
}

func thousands(value string) string {
	for i := len(value) - 3; i > 0; i -= 3 {
		value = value[:i] + "," + value[i:]
	}
	return value
}

func compact(value int64) string {
	v := float64(value)
	for _, unit := range []string{"", "K", "M", "B", "T"} {
		if v < 1000 || unit == "T" {
			switch {
			case v >= 100:
				return fmt.Sprintf("%.0f%s", v, unit)
			case v >= 10:
				return fmt.Sprintf("%.1f%s", v, unit)
			default:
				return fmt.Sprintf("%.2f%s", v, unit)
			}
		}
		v /= 1000
	}
	return fmt.Sprintf("%.0f", v)
}

func percent(value float64) string {
	return fmt.Sprintf("%.1f%%", value*100)
}

func plural(value int, noun string) string {
	if value == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", value, noun)
}
