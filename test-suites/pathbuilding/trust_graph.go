package pathbuilding

import "sort"

// A TrustGraph is abstractly a directed graph (potentially with cycles). It represents the trust relationship between
// entities where are arrow represents a certificate signed by the source entity for the destination entity. A
// TrustGraph can also label some edges as "invalid" meaning that there is a certificate but it should be considered
// invalid, e.g. because it is expired.
type TrustGraph struct {
	name  string
	nodes []string
	edges []Edge
}

// An Edge in a TrustGraph
type Edge struct {
	Source      string
	Destination string
}

func (e *Edge) Equals(other *Edge) bool {
	return e.Source == other.Source && e.Destination == other.Destination
}
func (e *Edge) MemberOf(s []Edge) bool {
	for _, other := range s {
		if e.Equals(&other) {
			return true
		}
	}
	return false
}

// NewGraph creates a TrustGraph instance with the given edges, where all edges are considered valid
func NewGraph(name string, edges []Edge) *TrustGraph {
	nodeNames := NewStringSet()
	for _, edge := range edges {
		nodeNames.Add(edge.Source)
		nodeNames.Add(edge.Destination)
	}
	nodes := nodeNames.Values()
	sort.Strings(nodes)

	gEdges := make([]Edge, len(edges))
	copy(gEdges, edges)
	return &TrustGraph{
		name:  name,
		nodes: nodes,
		edges: gEdges,
	}
}

func (g *TrustGraph) Name() string {
	return g.name
}

// NodeNames returns a slice of all the names of nodes in the graph
func (g *TrustGraph) NodeNames() []string {
	return g.nodes
}

// EdgeCount returns the number of edges in the graph (including both valid and invalid edges)
func (g *TrustGraph) EdgeCount() uint {
	return uint(len(g.edges))
}

// GetAllEdges returns the edges from this graph
func (g *TrustGraph) GetAllEdges() []Edge {
	res := make([]Edge, len(g.edges))
	copy(res, g.edges)
	return res
}

func stringInSlice(haystack []string, needle string) bool {
	for _, val := range haystack {
		if val == needle {
			return true
		}
	}
	return false
}

// Reachable returns a path if there is a path in the graph from the src node to the dst node, following only valid
// edges. If there is no path, this returns nil.
func (g *TrustGraph) Reachable(invalidEdges []Edge, src string, dst string) []string {
	var dfsIterate func(path []string, start string) []string
	dfsIterate = func(path []string, start string) []string {
		if stringInSlice(path, start) {
			return nil
		}

		newPath := append(path, start)
		if start == dst {
			return newPath
		}
		for _, edge := range g.edges {
			if edge.Source != start {
				continue
			}
			if edge.MemberOf(invalidEdges) {
				continue
			}

			nextNode := edge.Destination
			foundPath := dfsIterate(newPath, nextNode)
			if foundPath != nil {
				return foundPath
			}
		}
		return nil
	}
	return dfsIterate(nil, src)
}

var LINEAR_TRUST_GRAPH = NewGraph("LINEAR_TRUST_GRAPH", []Edge{
	{"ICA", "EE"},
	{"Trust Anchor", "ICA"},
})

/*
https://datatracker.ietf.org/doc/html/rfc4158#section-2.3

	  +---------+
	  |  Trust  |
	  | Anchor  |
	  +---------+
	   |       |
	   v       v
	+---+    +---+
	| A |<-->| C |
	+---+    +---+
	 |         |
	 |  +---+  |
	 +->| B |<-+
	    +---+
	      |
	      v
	    +----+
	    | EE |
	    +----+
*/
var FIGURE_SEVEN = NewGraph("FIGURE_SEVEN", []Edge{
	{"B", "EE"},
	{"C", "B"},
	{"A", "B"},
	{"C", "A"},
	{"A", "C"},
	{"Trust Anchor", "C"},
	{"Trust Anchor", "A"},
})

var TWO_ROOTS = NewGraph("TWO_ROOTS", []Edge{
	{"ICA", "EE"},
	{"Root1", "ICA"},
	{"Root2", "ICA"},
})

/*
https://datatracker.ietf.org/doc/html/rfc4158#section-2.4

	   +---+    +---+
	   | F |--->| H |
	   +---+    +---+
	    ^ ^       ^
	    |  \       \
	    |   \       \
	    |    v       v
	    |  +---+    +---+
	    |  | G |--->| I |
	    |  +---+    +---+
	    |   ^
	    |  /
	    | /
	+------+       +-----------+        +------+   +---+   +---+
	| TA W |<----->| Bridge CA |<------>| TA X |-->| L |-->| M |
	+------+       +-----------+        +------+   +---+   +---+
	                  ^      ^               \        \
	                 /        \               \        \
	                /          \               \        \
	               v            v               v        v
	         +------+         +------+        +---+    +---+
	         | TA Y |         | TA Z |        | J |    | N |
	         +------+         +------+        +---+    +---+
	          /   \              / \            |        |
	         /     \            /   \           |        |
	        /       \          /     \          v        v
	       v         v        v       v       +---+    +----+
	     +---+     +---+    +---+   +---+     | K |    | EE |
	     | A |<--->| C |    | O |   | P |     +---+    +----+
	     +---+     +---+    +---+   +---+
	        \         /      /  \       \
	         \       /      /    \       \
	          \     /      v      v       v
	           v   v    +---+    +---+   +---+
	           +---+    | Q |    | R |   | S |
	           | B |    +---+    +---+   +---+
	           +---+               |
	             /\                |
	            /  \               |
	           v    v              v
	        +---+  +---+         +---+
	        | E |  | D |         | T |
	        +---+  +---+         +---+
*/
var BRIDGE_CA_PKI = NewGraph("BRIDGE_CA_PKI", []Edge{
	{"F", "H"},
	{"F", "G"},
	{"G", "F"},
	{"H", "I"},
	{"I", "H"},
	{"G", "I"},
	{"TA W", "F"},
	{"TA W", "G"},

	{"J", "K"},
	{"N", "EE"},
	{"L", "N"},
	{"L", "M"},
	{"TA X", "J"},
	{"TA X", "L"},

	{"B", "E"},
	{"B", "D"},
	{"A", "B"},
	{"C", "B"},
	{"A", "C"},
	{"C", "A"},
	{"TA Y", "A"},
	{"TA Y", "C"},

	{"R", "S"},
	{"O", "R"},
	{"O", "Q"},
	{"P", "S"},
	{"TA Z", "O"},
	{"TA Z", "P"},

	{"TA W", "Bridge CA"},
	{"Bridge CA", "TA W"},
	{"TA X", "Bridge CA"},
	{"Bridge CA", "TA X"},
	{"TA Y", "Bridge CA"},
	{"Bridge CA", "TA Y"},
	{"TA Z", "Bridge CA"},
	{"Bridge CA", "TA Z"},
})

var ALL_TRUST_GRAPHS = []*TrustGraph{
	TWO_ROOTS,
	LINEAR_TRUST_GRAPH,
	FIGURE_SEVEN,
	BRIDGE_CA_PKI,
}
