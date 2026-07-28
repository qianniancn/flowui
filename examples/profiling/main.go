package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/components/checkbox"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	textui "github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	duration   = flag.Int("duration", 10, "profiling duration in seconds")
	frames     = flag.Int("frames", 0, "number of frames to render (0 = time-based)")
)

func main() {
	flag.Parse()
	stopProfile := startCPUProfile(*cpuprofile)

	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("FlowUI Profiling"),
			app.Size(unit.Dp(1024), unit.Dp(768)),
		)
		err := runProfiling(w)
		if stopProfile != nil {
			stopProfile()
		}
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func startCPUProfile(path string) func() {
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close()
		log.Fatal("could not start CPU profile: ", err)
	}
	return func() {
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			log.Print("could not close CPU profile: ", err)
		}
	}
}

type profilingState struct {
	ctx         *frame.Context
	frameCount  int
	startTime   time.Time
	checkStates [10]bool
}

func runProfiling(w *app.Window) error {
	activeTheme := theme.DefaultTheme()
	state := &profilingState{
		ctx:       frame.New(w, &activeTheme, locale.LanguageAuto),
		startTime: time.Now(),
	}

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			state.frameCount++
			state.layoutProfilingUI(gtx)
			e.Frame(gtx.Ops)

			if *frames > 0 && state.frameCount >= *frames {
				fmt.Printf("Rendered %d frames\n", state.frameCount)
				return nil
			}
			if *frames == 0 && time.Since(state.startTime) >= time.Duration(*duration)*time.Second {
				fps := float64(state.frameCount) / time.Since(state.startTime).Seconds()
				fmt.Printf("Rendered %d frames in %v (%.2f FPS)\n",
					state.frameCount, time.Since(state.startTime), fps)
				return nil
			}
			w.Invalidate()
		}
	}
}

func (s *profilingState) layoutProfilingUI(gtx layout.Context) layout.Dimensions {
	frame.BeginFrameWithViewport(s.ctx, gtx.Constraints.Max)
	dims := layoutui.Column(
		s.layoutHeader(),
		s.layoutControls(),
		s.layoutGrid(),
		s.layoutFooter(),
	).Layout(s.ctx, gtx)
	frame.LayoutOverlays(s.ctx, gtx)
	frame.ApplyFrameCommands(s.ctx, gtx)
	frame.EndFrame(s.ctx)
	return dims
}

func (s *profilingState) layoutHeader() frame.Widget {
	headerStyle := style.Style{}.
		Padding(16).
		Background(style.TokenSurface).
		BoxShadow(0, 2, 4, 0, style.WithAlpha(style.TokenSurfaceShadow, 0.1))
	return layoutui.Box(textui.New("FlowUI Profiling Test")).Style(headerStyle)
}

func (s *profilingState) layoutControls() frame.Widget {
	children := make([]frame.Widget, len(s.checkStates))
	for i := range children {
		idx := i
		children[i] = checkbox.Checkbox(fmt.Sprintf("check-%d", idx), s.checkStates[idx], fmt.Sprintf("Option %d", idx+1)).
			OnChange(func(value bool) { s.checkStates[idx] = value })
	}
	return layoutui.Wrap(children...).Gap(8)
}

func (s *profilingState) layoutGrid() frame.Widget {
	rows := make([]frame.Widget, 8)
	for row := range rows {
		cols := make([]frame.Widget, 6)
		for col := range cols {
			btnStyle := style.Style{}.
				Radius(unit.Dp(8)).
				Background(s.buttonColor(row, col)).
				When(style.Hovered, style.Style{}.
					Background(s.buttonHoverColor(row, col))).
				Transition(style.PropBackgroundColor, 150*time.Millisecond)
			cols[col] = button.Button(
				fmt.Sprintf("btn-%d-%d", row, col),
				textui.New(fmt.Sprintf("B%d", row*6+col+1)),
			).Style(btnStyle)
		}
		rows[row] = layoutui.Row(cols...).Gap(8)
	}
	return layoutui.Column(rows...).Gap(8)
}

func (s *profilingState) layoutFooter() frame.Widget {
	info := fmt.Sprintf("Frame: %d | Goroutines: %d", s.frameCount, runtime.NumGoroutine())
	footerStyle := style.Style{}.
		Padding(16).
		Background(style.TokenSurfaceSecondary).
		TextColor(style.TokenMutedForeground)
	return layoutui.Box(textui.New(info)).Style(footerStyle)
}

func (s *profilingState) buttonColor(row, col int) style.ThemeColor {
	colors := []style.ThemeColor{
		style.TokenAccent,
		style.TokenDefault,
		style.TokenSurfaceTertiary,
		style.TokenDanger,
		style.TokenSuccess,
		style.TokenSurface,
	}
	return colors[col%len(colors)]
}

func (s *profilingState) buttonHoverColor(row, col int) style.ThemeColor {
	colors := []style.ThemeColor{
		style.TokenAccentHover,
		style.TokenDefaultHover,
		style.TokenSurfaceHover,
		style.TokenDangerHover,
		style.TokenSuccessSoft,
		style.TokenSurfaceHover,
	}
	return colors[col%len(colors)]
}
