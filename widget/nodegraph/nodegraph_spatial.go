package nodegraph

import (
	"math"
	"sort"
)

const (
	graphSpatialCellSize = float32(256)
	graphSpatialMaxCells = 256
)

type graphSpatialCell struct{ x, y int }

type graphSpatialIndex struct {
	nodes        map[graphSpatialCell][]int
	nodeFallback []int
	edges        map[graphSpatialCell][]int
	edgeFallback []int
	nodeCount    int
	edgeCount    int
}

func (g *resolvedGraph) buildSpatialIndex() {
	index := graphSpatialIndex{
		nodes:     make(map[graphSpatialCell][]int),
		edges:     make(map[graphSpatialCell][]int),
		nodeCount: len(g.nodes),
		edgeCount: len(g.edges),
	}
	for position, node := range g.nodes {
		bounds := graphRect{min: node.position, max: Point{X: node.position.X + node.size.Width, Y: node.position.Y + node.size.Height}}
		if !index.add(index.nodes, bounds, position) {
			index.nodeFallback = append(index.nodeFallback, position)
		}
	}
	for position, edge := range g.edges {
		padding := graphEdgeCullingPadding(edge.source, edge.target)
		bounds := graphRect{min: Point{X: min(edge.source.X, edge.target.X) - padding, Y: min(edge.source.Y, edge.target.Y) - padding}, max: Point{X: max(edge.source.X, edge.target.X) + padding, Y: max(edge.source.Y, edge.target.Y) + padding}}
		if !index.add(index.edges, bounds, position) {
			index.edgeFallback = append(index.edgeFallback, position)
		}
	}
	g.spatial = index
}

func (i graphSpatialIndex) add(cells map[graphSpatialCell][]int, rect graphRect, value int) bool {
	minCell, maxCell := spatialCellFor(rect.min), spatialCellFor(rect.max)
	width, height := maxCell.x-minCell.x+1, maxCell.y-minCell.y+1
	if width <= 0 || height <= 0 || width > graphSpatialMaxCells || height > graphSpatialMaxCells || width*height > graphSpatialMaxCells {
		return false
	}
	for y := minCell.y; y <= maxCell.y; y++ {
		for x := minCell.x; x <= maxCell.x; x++ {
			cell := graphSpatialCell{x: x, y: y}
			cells[cell] = append(cells[cell], value)
		}
	}
	return true
}

func (i graphSpatialIndex) nodeCandidates(rect graphRect) []int {
	return i.candidates(i.nodes, i.nodeFallback, i.nodeCount, rect)
}

func (i graphSpatialIndex) edgeCandidates(rect graphRect) []int {
	return i.candidates(i.edges, i.edgeFallback, i.edgeCount, rect)
}

func (i graphSpatialIndex) candidates(cells map[graphSpatialCell][]int, fallback []int, count int, rect graphRect) []int {
	if count == 0 {
		return nil
	}
	seen := make(map[int]struct{})
	minCell, maxCell := spatialCellFor(rect.min), spatialCellFor(rect.max)
	for y := minCell.y; y <= maxCell.y; y++ {
		for x := minCell.x; x <= maxCell.x; x++ {
			for _, value := range cells[graphSpatialCell{x: x, y: y}] {
				seen[value] = struct{}{}
			}
		}
	}
	for _, value := range fallback {
		seen[value] = struct{}{}
	}
	result := make([]int, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result
}

func spatialCellFor(point Point) graphSpatialCell {
	return graphSpatialCell{x: int(math.Floor(float64(point.X / graphSpatialCellSize))), y: int(math.Floor(float64(point.Y / graphSpatialCellSize)))}
}
