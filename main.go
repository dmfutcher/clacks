package main

import (
	"fmt"
	"time"

	"github.com/hippoai/graphgo"
	zmq "github.com/pebbe/zmq4"
)

func zmqAddress(station string) string {
	return fmt.Sprintf("ipc://%v", station)
}

type Station struct {
	name       string
	socket     *zmq.Socket
	graph_node *graphgo.Node
	graph      *graphgo.Graph
}

func newStation(name string, node *graphgo.Node, graph *graphgo.Graph) *Station {
	station := new(Station)
	station.name = name
	station.graph_node = node
	station.graph = graph

	sock, _ := zmq.NewSocket(zmq.REP)
	sock.Bind(zmqAddress(name))

	station.socket = sock
	return station
}

func (station *Station) serve() {
	station.log("Listening")
	for {
		msg, _ := station.socket.Recv(0)
		station.log(fmt.Sprint("RECEIVED: ", msg))
		station.socket.Send("ACK", 0) // TODO: Use different ZeroMQ socket type to avoid insta-ack?

		station.publish(msg) // TODO: Smarter publish / drop logic to deal with infinite loops
	}
}

func (station *Station) publish(msg string) {
	peer_edges, err := station.graph_node.OutE(station.graph, "CONNECTS")
	if err != nil {
		fmt.Println("Failed to find neighbour stations for ", station.name)
	}

	for _, edge := range peer_edges {
		start_node, err := edge.EndN(station.graph)
		if err != nil {
			fmt.Println("Failed to find neighbour stations from neighbour edge ", station.name)
		}

		peer_label := start_node.GetKey()
		requester, _ := zmq.NewSocket(zmq.REQ)
		defer requester.Close()
		requester.Connect(zmqAddress(peer_label))
		requester.Send(msg, 0)
		station.log(fmt.Sprint("SENT: ", msg, " to ", peer_label))
	}

}

func (station *Station) log(msg string) {
	fmt.Println("[", station.name, "]", msg)
}

func createStationNode(g *graphgo.Graph, label string, peer_station_labels []string) (*graphgo.Node, error) {
	station := newStation(label, nil, g)
	node, err := g.MergeNode(label, map[string]interface{}{"station": station})
	if err != nil {
		return nil, err
	}

	station.graph_node = node

	for _, peer_label := range peer_station_labels {
		g.MergeEdge(
			fmt.Sprint("connect.", label, ".", peer_label), "CONNECTS",
			label, peer_label,
			map[string]interface{}{},
		)

		g.MergeEdge(
			fmt.Sprint("connect.", peer_label, ".", label), "CONNECTS",
			peer_label, label,
			map[string]interface{}{},
		)
	}

	go station.serve()

	return node, nil
}

func newStationGraph(station_defs *map[string][]string) *graphgo.Graph {
	g := graphgo.NewEmptyGraph()
	for station, peers := range *station_defs {
		createStationNode(g, station, peers)
	}

	return g
}

func main() {
	station_defs := map[string][]string{
		"station.0": []string{"station.1"},
		"station.1": []string{"station.0"},
	}

	graph := newStationGraph(&station_defs)

	station_0_node, _ := graph.GetNode("station.0")
	node_station, err := station_0_node.Get("station")
	var station0 *Station
	if err == nil {
		station0 = node_station.(*Station)
	} else {
		fmt.Println("Failed to get station.0")
	}
	time.Sleep(2 * time.Second)

	station0.publish("TEST MESSAGE")

	time.Sleep(30 * time.Second)
}
