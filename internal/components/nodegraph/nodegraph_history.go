package nodegraph

// Fragment is a portable, application-owned selection of nodes and the edges
// that connect only those nodes. It is suitable for copy, cut, and paste
// commands. Custom node content is retained in-memory; applications that need
// cross-process clipboard data should map their own serializable node data.
type Fragment struct {
	Nodes []Node
	Edges []Edge
}

// CopySelection returns the selected nodes, all selected descendants, and
// their internal edges. A child copied without its parent becomes a root node
// at its current world position, so the result is always a valid graph.
func CopySelection(graph Graph, selected map[string]bool) Fragment {
	resolved := resolveGraph(graph)
	included := make(map[string]bool, len(selected))
	for id := range selected {
		if _, exists := resolved.byID[id]; exists {
			included[id] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for _, node := range resolved.nodes {
			if node.node.parentID != "" && included[node.node.parentID] && !included[node.node.ID] {
				included[node.node.ID] = true
				changed = true
			}
		}
	}
	fragment := Fragment{Nodes: make([]Node, 0, len(included))}
	for _, resolvedNode := range resolved.nodes {
		if !included[resolvedNode.node.ID] {
			continue
		}
		node := cloneNode(resolvedNode.node)
		if node.parentID != "" && !included[node.parentID] {
			node.parentID = ""
			node.Position = resolvedNode.position
		}
		fragment.Nodes = append(fragment.Nodes, node)
	}
	for _, edge := range graph.Edges {
		if included[edge.Source.NodeID] && included[edge.Target.NodeID] {
			fragment.Edges = append(fragment.Edges, edge)
		}
	}
	return fragment
}

// PasteFragment appends a fragment to graph using application-supplied stable
// IDs. Offset applies only to pasted roots, preserving each copied group's
// local child coordinates. It panics when either allocator is nil or returns
// an empty or duplicate ID.
func PasteFragment(graph Graph, fragment Fragment, nodeID func(Node) string, edgeID func(Edge) string, offset Point) Graph {
	if nodeID == nil || edgeID == nil {
		panic("flowui: node graph paste requires node and edge ID allocators")
	}
	result := Graph{Nodes: cloneNodes(graph.Nodes), Edges: cloneEdges(graph.Edges)}
	usedNodes := make(map[string]bool, len(result.Nodes)+len(fragment.Nodes))
	for _, node := range result.Nodes {
		usedNodes[node.ID] = true
	}
	remap := make(map[string]string, len(fragment.Nodes))
	for _, source := range fragment.Nodes {
		id := nodeID(source)
		if id == "" || usedNodes[id] {
			panic("flowui: node graph paste allocator returned an empty or duplicate node ID")
		}
		usedNodes[id] = true
		remap[source.ID] = id
	}
	for _, source := range fragment.Nodes {
		node := cloneNode(source)
		node.ID = remap[source.ID]
		if node.parentID == "" {
			node.Position.X += offset.X
			node.Position.Y += offset.Y
		} else {
			node.parentID = remap[node.parentID]
		}
		result.Nodes = append(result.Nodes, node)
	}
	usedEdges := make(map[string]bool, len(result.Edges)+len(fragment.Edges))
	for _, edge := range result.Edges {
		usedEdges[edge.ID] = true
	}
	for _, source := range fragment.Edges {
		edge := source
		edge.ID = edgeID(source)
		if edge.ID == "" || usedEdges[edge.ID] {
			panic("flowui: node graph paste allocator returned an empty or duplicate edge ID")
		}
		usedEdges[edge.ID] = true
		edge.Source.NodeID = remap[source.Source.NodeID]
		edge.Target.NodeID = remap[source.Target.NodeID]
		result.Edges = append(result.Edges, edge)
	}
	return result
}

// History stores application-owned graph snapshots for undo and redo. Commit
// only completed user operations, such as a drag release, to avoid recording
// every intermediate controlled change.
type History struct {
	entries []Graph
	index   int
	limit   int
}

// NewHistory creates a history initialized with graph. The default limit is
// 100 snapshots.
func NewHistory(graph Graph) *History {
	return &History{entries: []Graph{cloneGraph(graph)}, limit: 100}
}

// Limit sets the maximum retained snapshot count. It must be positive.
func (h *History) Limit(limit int) *History {
	if limit <= 0 {
		panic("flowui: node graph history limit must be positive")
	}
	h.limit = limit
	h.trim()
	return h
}

// Commit appends graph and discards redo snapshots. It returns false when h is
// nil and otherwise records the snapshot.
func (h *History) Commit(graph Graph) bool {
	if h == nil {
		return false
	}
	h.entries = append(h.entries[:h.index+1], cloneGraph(graph))
	h.index = len(h.entries) - 1
	h.trim()
	return true
}

// CanUndo reports whether Undo can return an earlier snapshot.
func (h *History) CanUndo() bool {
	return h != nil && h.index > 0
}

// CanRedo reports whether Redo can return a later snapshot.
func (h *History) CanRedo() bool {
	return h != nil && h.index+1 < len(h.entries)
}

// Undo returns the preceding snapshot without exposing internal slice storage.
func (h *History) Undo() (Graph, bool) {
	if !h.CanUndo() {
		return Graph{}, false
	}
	h.index--
	return cloneGraph(h.entries[h.index]), true
}

// Redo returns the next snapshot without exposing internal slice storage.
func (h *History) Redo() (Graph, bool) {
	if !h.CanRedo() {
		return Graph{}, false
	}
	h.index++
	return cloneGraph(h.entries[h.index]), true
}

func (h *History) trim() {
	if h == nil || h.limit <= 0 || len(h.entries) <= h.limit {
		return
	}
	drop := len(h.entries) - h.limit
	h.entries = append([]Graph(nil), h.entries[drop:]...)
	h.index = max(h.index-drop, 0)
}

func cloneGraph(graph Graph) Graph {
	return Graph{Nodes: cloneNodes(graph.Nodes), Edges: cloneEdges(graph.Edges)}
}

func cloneNodes(nodes []Node) []Node {
	cloned := make([]Node, len(nodes))
	for index, node := range nodes {
		cloned[index] = cloneNode(node)
	}
	return cloned
}

func cloneNode(node Node) Node {
	node.InputPorts = append([]Port(nil), node.InputPorts...)
	node.OutputPorts = append([]Port(nil), node.OutputPorts...)
	return node
}

func cloneEdges(edges []Edge) []Edge {
	return append([]Edge(nil), edges...)
}
