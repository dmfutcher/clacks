package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"futcher.io/clacks/schema"

	"github.com/hippoai/graphgo"
	zmq "github.com/pebbe/zmq4"
	"google.golang.org/protobuf/proto"
)

func zmqAddress(station string) string {
	return fmt.Sprintf("ipc://%v", station)
}

type Station struct {
	name      string
	socket    *zmq.Socket
	graphNode *graphgo.Node
	graph     *graphgo.Graph
}

func newStation(name string, node *graphgo.Node, graph *graphgo.Graph) *Station {
	station := new(Station)
	station.name = name
	station.graphNode = node
	station.graph = graph

	sock, _ := zmq.NewSocket(zmq.REP)
	sock.Bind(zmqAddress(name))

	station.socket = sock
	return station
}

func (station *Station) serve() {
	station.log("Listening")
	for {
		wireFrame, _ := station.socket.RecvBytes(0)
		station.socket.Send("ACK", 0) // TODO: Use different ZeroMQ socket type to avoid insta-ack?

		frame := &schema.Frame{}
		if err := proto.Unmarshal(wireFrame, frame); err != nil {
			station.log("Failed to unmarshal message from wire format")
		}

		station.log(fmt.Sprint("RECEIVED: ", frame))
		station.relay(frame)
	}
}

func (station *Station) relay(frame *schema.Frame) {
	peerEdges, err := station.graphNode.OutE(station.graph, "CONNECTS")
	if err != nil {
		fmt.Println("Failed to find neighbour stations for ", station.name)
	}

	for _, edge := range peerEdges {
		startNode, err := edge.EndN(station.graph)
		if err != nil {
			fmt.Println("Failed to find neighbour stations from neighbour edge ", station.name)
		}

		peerLabel := startNode.GetKey()

		// Skip propagating message to the orgin or immediate predecessor stations
		if peerLabel == frame.Source || peerLabel == frame.Referrer {
			continue
		}

		requester, _ := zmq.NewSocket(zmq.REQ)
		defer requester.Close()
		requester.Connect(zmqAddress(peerLabel))

		wireFormatFrame, _ := proto.Marshal(frame)
		requester.SendBytes(wireFormatFrame, 0)
		station.log(fmt.Sprint("SENT: ", frame, " to ", peerLabel))
	}

}

func (station *Station) publish(body string) {
	frame := station.createMessage(body)
	station.relay(frame)
}

func (station *Station) createMessage(body string) *schema.Frame {
	hasher := sha256.New()
	hasher.Write([]byte(body))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return &schema.Frame{
		Hash:   hash,
		Source: station.name,
		Body:   body,
	}
}

func (station *Station) log(msg string) {
	fmt.Println("[", station.name, "]", msg)
}

func createStationNode(g *graphgo.Graph, label string, peerStationLabels []string) (*graphgo.Node, error) {
	station := newStation(label, nil, g)
	node, err := g.MergeNode(label, map[string]interface{}{"station": station})
	if err != nil {
		return nil, err
	}

	station.graphNode = node

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

	go station.serve()

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
		"station.0": {"station.1"},
		"station.1": {"station.0"},
		"station.2": {"station.1"},
	}

	graph := newStationGraph(&stationDefs)

	station0Node, _ := graph.GetNode("station.0")
	nodeStation, err := station0Node.Get("station")
	var station0 *Station
	if err == nil {
		station0 = nodeStation.(*Station)
	} else {
		fmt.Println("Failed to get station.0")
	}
	time.Sleep(2 * time.Second)

	station0.publish("TEST MESSAGE")

	time.Sleep(1 * time.Second)

}
