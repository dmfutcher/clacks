package main

import (
	"fmt"
	"strconv"
	"time"

	zmq "github.com/pebbe/zmq4"
)

// func client(freqs []int) {
// 	time.Sleep(time.Second * 3)
// 	fmt.Println("Connecting to hello world server...")
// 	requester, _ := zmq.NewSocket(zmq.REQ)
// 	defer requester.Close()
// 	requester.Connect("tcp://localhost:5555")

// 	for request_nbr := 0; request_nbr != 10; request_nbr++ {
// 		// send hello
// 		msg := fmt.Sprintf("Hello %d", request_nbr)
// 		fmt.Println("Sending ", msg)
// 		requester.Send(msg, 0)

// 		// Wait for reply:
// 		reply, _ := requester.Recv(0)
// 		fmt.Println("Received ", reply)
// 	}
// }

type Station struct {
	name             string
	socket           *zmq.Socket
	connection_freqs []int
}

func newStation(name string, frequency int, connections []int) *Station {
	station := new(Station)
	station.name = name
	station.connection_freqs = connections

	sock, _ := zmq.NewSocket(zmq.REP)
	sock.Bind(fmt.Sprintf("ipc://%v", strconv.Itoa(frequency)))

	station.socket = sock
	return station
}

func (station *Station) serve() {
	station.log("Listening")
	for {
		//  Wait for next request from client
		msg, _ := station.socket.Recv(0)
		station.log(fmt.Sprint("Received ", msg))
		station.socket.Send("ACK", 0)
	}
}

func (station *Station) publish(msg string) {
	for _, freq := range station.connection_freqs {
		station.log(fmt.Sprint("Publishing: ", msg, " to ", freq))

		requester, _ := zmq.NewSocket(zmq.REQ)
		defer requester.Close()
		requester.Connect(fmt.Sprint("ipc://", strconv.Itoa(freq)))
		requester.Send(msg, 0)
	}

}

func (station *Station) log(msg string) {
	fmt.Println("[", station.name, "] ", msg)
}

func main() {

	station := newStation("0", 100, []int{200})
	station_two := newStation("1", 200, []int{100, 300})
	station_three := newStation("2", 300, []int{200})

	go station.serve()
	go station_two.serve()
	go station_three.serve()

	time.Sleep(2 * time.Second)

	station.publish("test")
	// go station_three.serve()

	time.Sleep(30 * time.Second)

}
