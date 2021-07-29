package main

import (
	"fmt"
	"time"

	"futcher.io/clacks/station"

	"github.com/hippoai/graphgo"
)

type Network struct {
	graph *graphgo.Graph
}

func (n *Network) GetStation(label string) *station.Station {
	stationNode, _ := n.graph.GetNode(label)
	nodeStation, err := stationNode.Get("station")

	if err == nil {
		return nodeStation.(*station.Station)
	} else {
		return nil
	}
}

func (n *Network) String() string {
	return fmt.Sprintf("%v\n\n%v", n.graph.Nodes, n.graph.Edges)

}

func createStationNode(g *graphgo.Graph, label string) (*graphgo.Node, error) {
	station := station.New(label, g)
	node, err := g.MergeNode(label, map[string]interface{}{"station": station})
	if err != nil {
		fmt.Println("Could not create node for ", station, err)
		return nil, err
	}

	station.SetGraphNode(node)

	go station.Serve()

	return node, nil
}

func createStationConnections(g *graphgo.Graph, label string, peerStationLabels []string) {
	for _, peerLabel := range peerStationLabels {
		g.MergeEdge(
			fmt.Sprint("connect.", label, ".", peerLabel), "CONNECTS",
			label, peerLabel,
			map[string]interface{}{},
		)

		g.MergeEdge(
			fmt.Sprint("connect.", peerLabel, ".", label), "CONNECTS",
			peerLabel, label,
			map[string]interface{}{},
		)
	}
}

func newNetwork(stationDefs *map[string][]string) *Network {
	g := graphgo.NewEmptyGraph()
	for station := range *stationDefs {
		createStationNode(g, station)
	}

	for station, peerLabels := range *stationDefs {
		createStationConnections(g, station, peerLabels)
	}

	fmt.Println("Created graph: ", len(g.Nodes), "Vertices/", len(g.Edges), "Edges")

	return &Network{
		graph: g,
	}
}

func main() {
	stationDefs := map[string][]string{
		"station.0":   {"station.1"},
		"station.1":   {"station.0"},
		"station.1.1": {"station.1"},
		"station.2":   {"station.1"},
		"station.3.0": {"station.2"},
		"station.3.1": {"station.3.0"},
		"station.3.2": {"station.3.1"},
	}

	network := newNetwork(&stationDefs)

	time.Sleep(3 * time.Second)
	station := network.GetStation("station.2")

	for i := 0; i < 10; i++ {
		go station.Publish(fmt.Sprint("TEST MESSAGE", i))
	}

	time.Sleep(5 * time.Second)

	for station := range stationDefs {
		s := network.GetStation(station)
		fmt.Println(station, s.Drops(), len(s.Drops()))
	}

}
