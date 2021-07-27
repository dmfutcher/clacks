package main

import (
	"fmt"
	"time"

	"futcher.io/clacks/station"

	"github.com/hippoai/graphgo"
)

func createStationNode(g *graphgo.Graph, label string, peerStationLabels []string) (*graphgo.Node, error) {
	station := station.New(label, nil, g)
	node, err := g.MergeNode(label, map[string]interface{}{"station": station})
	if err != nil {
		return nil, err
	}

	station.SetGraphNode(node)

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

	go station.Serve()

	return node, nil
}

func newStationGraph(stationDefs *map[string][]string) *graphgo.Graph {
	g := graphgo.NewEmptyGraph()
	for station, peers := range *stationDefs {
		createStationNode(g, station, peers)
	}

	return g
}

func main() {
	stationDefs := map[string][]string{
		"station.0":   {"station.1", "station.3.2"},
		"station.1":   {"station.0"},
		"station.1.1": {"station.1"},
		"station.2":   {"station.1"},
		"station.3.0": {"station.1"},
		"station.3.1": {"station.3.0"},
		"station.3.2": {"station.3.1", "station.0"},
	}

	graph := newStationGraph(&stationDefs)

	station0Node, _ := graph.GetNode("station.0")
	nodeStation, err := station0Node.Get("station")
	var station0 *station.Station
	if err == nil {
		station0 = nodeStation.(*station.Station)
	} else {
		fmt.Println("Failed to get station.0")
	}
	time.Sleep(2 * time.Second)

	station0.Publish("TEST MESSAGE")

	time.Sleep(5 * time.Second)

}
