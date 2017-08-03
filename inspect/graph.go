package inspect

import (
	"github.com/quilt/quilt/stitch"
)

// A Node in the communiction Graph.
type Node struct {
	Name        string
	Label       string
	Connections map[string]Node
}

// An Edge in the communication Graph.
type Edge struct {
	From string
	To   string
}

// A Graph represents permission to communicate across a series of Nodes.
// Each Node is a container and each edge is permissions to
// initiate a connection.
type Graph struct {
	Nodes map[string]Node
}

// New queries the Stitch to create a Graph structure.
func New(blueprint stitch.Stitch) (Graph, error) {
	g := Graph{
		Nodes: map[string]Node{},
	}

	for _, label := range blueprint.Labels {
		for _, cid := range label.IDs {
			g.addNode(cid, label.Name)
		}
	}
	g.addNode(stitch.PublicInternetLabel, stitch.PublicInternetLabel)

	for _, conn := range blueprint.Connections {
		err := g.addConnection(conn.From, conn.To)
		if err != nil {
			return Graph{}, err
		}
	}

	return g, nil
}

// GetConnections returns a list of the edges in the Graph.
func (g Graph) GetConnections() []Edge {
	var res []Edge
	for _, n := range g.Nodes {
		for _, edge := range n.Connections {
			res = append(res, Edge{From: n.Name, To: edge.Name})
		}
	}
	return res
}

func (g *Graph) addConnection(from string, to string) error {
	// from and to are labels
	var fromContainers []Node
	var toContainers []Node

	for _, node := range g.Nodes {
		if node.Label == from {
			fromContainers = append(fromContainers, node)
		}
		if node.Label == to {
			toContainers = append(toContainers, node)
		}
	}

	for _, fromNode := range fromContainers {
		for _, toNode := range toContainers {
			if fromNode.Name != toNode.Name {
				fromNode.Connections[toNode.Name] = toNode
			}
		}
	}

	return nil
}

func (g *Graph) addNode(cid string, label string) Node {
	n := Node{
		Name:        cid,
		Label:       label,
		Connections: map[string]Node{},
	}
	g.Nodes[cid] = n

	return n
}
