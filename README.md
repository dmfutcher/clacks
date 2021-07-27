= clacks =
*clacks* is toy a simulation of a very simple message relaying network. 'Stations' are vertices on a graph, with edges between them representing interconnections. A message can be published by a station that passes it to all their peer stations, who in turn pass it to their peer stations. Entirely built so I can shake off some rust with Golang and play with a couple libraries. *clacks* uses Graphgo for the station graph, ZeroMQ for message transport and Protobuf for serialisation. 



== Run it ==

```
brew install zeromq protobuf
protoc --go_out=. message.proto
go run main.go
```