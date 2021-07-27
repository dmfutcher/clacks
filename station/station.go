package station

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"futcher.io/clacks/schema"
	"github.com/hippoai/graphgo"
	zmq "github.com/pebbe/zmq4"
	ring "github.com/zealws/golang-ring"
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
	drops     *ring.Ring
}

func New(name string, node *graphgo.Node, graph *graphgo.Graph) *Station {
	station := &Station{
		name:      name,
		graphNode: node,
		graph:     graph,
	}

	sock, _ := zmq.NewSocket(zmq.REP)
	sock.Bind(zmqAddress(name))
	station.socket = sock

	station.drops = &ring.Ring{}
	station.drops.SetCapacity(10)

	return station
}

func (station *Station) Serve() {
	station.log("Listening")
	for {
		wireFrame, _ := station.socket.RecvBytes(0)
		station.socket.Send("ACK", 0) // TODO: Use different ZeroMQ socket type to avoid insta-ack?

		frame := &schema.Frame{}
		if err := proto.Unmarshal(wireFrame, frame); err != nil {
			station.log("Failed to unmarshal message from wire format")
		}

		station.relay(frame)
	}
}

func (station *Station) relay(frame *schema.Frame) {
	// Drop messages in the "recently seen" "drop list"
	for _, hash := range station.drops.Values() {
		if hash == frame.Hash {
			return
		}
	}
	station.drops.Enqueue(frame.Hash)

	station.log(fmt.Sprint("RECEIVED: ", frame))

	peerEdges, err := station.graphNode.OutE(station.graph, "CONNECTS")
	if err != nil {
		fmt.Println("Failed to find neighbour stations for ", station.name)
	}

	originalReferrer := frame.Referrer
	frame.Referrer = station.name
	frame.Hops += 1

	for _, edge := range peerEdges {
		startNode, err := edge.EndN(station.graph)
		if err != nil {
			fmt.Println("Failed to find neighbour stations from neighbour edge ", station.name)
		}

		peerLabel := startNode.GetKey()

		// Skip propagating message to the orgin or immediate predecessor stations
		if peerLabel == frame.Source || peerLabel == originalReferrer {
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

func (station *Station) Publish(body string) {
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
		Hops:   0,
	}
}

func (station *Station) log(msg string) {
	fmt.Println("[", station.name, "]", msg)
}

func (station *Station) SetGraphNode(node *graphgo.Node) {
	station.graphNode = node
}
