package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/barchart"
	"github.com/qianniancn/flowui/widget/candlestick"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/gantt"
	"github.com/qianniancn/flowui/widget/heatmap"
	"github.com/qianniancn/flowui/widget/linechart"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/piechart"
)

func chartSelection(send ui.Send[Msg]) func(chart.Selection) {
	return func(value chart.Selection) {
		label := value.Label
		if len(value.Items) > 0 {
			item := value.Items[0]
			label += fmt.Sprintf(" · %s：%.2f", item.SeriesLabel, item.Y)
		}
		send(logMsg("图表选择：" + label))
	}
}

func demoLine(_ demoState, send ui.Send[Msg]) ui.Widget {
	return linechart.New("line-demo", []linechart.Series{
		linechart.Values("visits", "访问量", []float64{28, 42, 36, 58, 64, 52, 78}).Smooth(true).Area(true),
		linechart.Values("signups", "注册量", []float64{12, 18, 16, 28, 32, 26, 40}).Smooth(true),
	}).Categories([]string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}).
		Height(300).Legend(true).OnDataClick(chartSelection(send))
}

func demoBar(_ demoState, send ui.Send[Msg]) ui.Widget {
	return barchart.New("bar-demo", []barchart.Series{
		barchart.Values("planned", "计划", []float64{40, 60, 48, 72}),
		barchart.Values("actual", "实际", []float64{36, 64, 44, 68}),
	}).Categories([]string{"第一周", "第二周", "第三周", "第四周"}).
		Height(300).Legend(true).OnDataClick(chartSelection(send))
}

func demoPie(_ demoState, send ui.Send[Msg]) ui.Widget {
	return piechart.New("pie-demo", []piechart.Data{
		piechart.Slice("desktop", "桌面", 56), piechart.Slice("web", "Web", 28),
		piechart.Slice("mobile", "移动端", 16),
	}).InnerRadius(0.3).Height(300).Legend(true).OnDataClick(chartSelection(send))
}

func demoCandle(_ demoState, send ui.Send[Msg]) ui.Widget {
	data := []candlestick.Candle{}
	for index := range 24 {
		open := 30 + float64(index)*0.6 + float64(index%3)*2
		close := open + float64(index%5) - 2
		data = append(data, candlestick.OHLC(open, close, min(open, close)-2, max(open, close)+3))
	}
	return candlestick.New("candle-demo", data).Height(300).OnDataClick(chartSelection(send))
}

func demoHeatmap(_ demoState, send ui.Send[Msg]) ui.Widget {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 4, 0)
	values := []heatmap.CalendarValue{}
	for day := range 122 {
		values = append(values, heatmap.CalendarValue{Date: start.AddDate(0, 0, day), Value: float64(day * 7 % 13)})
	}
	return heatmap.New("heatmap-demo", start, end, values).Height(240).OnDataClick(chartSelection(send))
}

func demoGantt(_ demoState, send ui.Send[Msg]) ui.Widget {
	start := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local)
	return gantt.New("gantt-demo", []gantt.Task{
		gantt.NewTask("design", "设计", start, start.AddDate(0, 0, 4)).Progress(1),
		gantt.NewTask("build", "开发", start.AddDate(0, 0, 3), start.AddDate(0, 0, 12)).Progress(0.65).DependsOn("design"),
		gantt.NewTask("test", "测试", start.AddDate(0, 0, 10), start.AddDate(0, 0, 16)).Progress(0.2).DependsOn("build"),
		gantt.NewTask("release", "发布", start.AddDate(0, 0, 16), start.AddDate(0, 0, 18)).DependsOn("test"),
	}).Height(300).Legend(true).OnTaskClick(func(value gantt.Selection) {
		send(logMsg(fmt.Sprintf("%s · %s 至 %s · 进度 %.0f%%",
			value.Label,
			value.Start.Format("01-02"),
			value.End.Format("01-02"),
			value.Progress*100,
		)))
	})
}

func initialGraph() nodegraph.Graph {
	return nodegraph.Graph{
		Nodes: []nodegraph.Node{
			nodegraph.NewNode("source", "数据源", nodegraph.Point{X: 20, Y: 40}).
				Outputs(nodegraph.NewPort("out", "数据")),
			nodegraph.NewNode("transform", "转换", nodegraph.Point{X: 250, Y: 120}).
				Inputs(nodegraph.NewPort("in", "输入")).Outputs(nodegraph.NewPort("out", "输出")),
			nodegraph.NewNode("result", "预览", nodegraph.Point{X: 480, Y: 40}).
				Inputs(nodegraph.NewPort("in", "结果")),
		},
		Edges: []nodegraph.Edge{
			nodegraph.NewEdge("source-transform", nodegraph.NewEndpoint("source", "out"), nodegraph.NewEndpoint("transform", "in")),
			nodegraph.NewEdge("transform-result", nodegraph.NewEndpoint("transform", "out"), nodegraph.NewEndpoint("result", "in")),
		},
	}
}

func demoGraph(m demoState, send ui.Send[Msg]) ui.Widget {
	return nodegraph.New("graph-demo", m.graph).Height(360).Minimap(true).
		OnNodesChange(func(changes []nodegraph.NodeChange) {
			send(demoMsg(func(m *demoState) { m.graph.Nodes = nodegraph.ApplyNodeChanges(m.graph.Nodes, changes) }))
		}).OnEdgesChange(func(changes []nodegraph.EdgeChange) {
		send(demoMsg(func(m *demoState) { m.graph.Edges = nodegraph.ApplyEdgeChanges(m.graph.Edges, changes) }))
	}).OnConnect(func(connection nodegraph.Connection) {
		send(demoMsg(func(m *demoState) {
			for _, edge := range m.graph.Edges {
				if edge.Source == connection.Source && edge.Target == connection.Target {
					return
				}
			}
			m.edgeID++
			id := fmt.Sprintf("edge-%d", m.edgeID)
			m.graph.Edges = append(m.graph.Edges, nodegraph.NewEdge(id, connection.Source, connection.Target))
		}))
	})
}
