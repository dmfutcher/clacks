clacks
======

**This is not a place of honour. There is no value to be found here**. This is *enirely* built to shake off some rust with Golang and play with some libraries. Available here on the miniscule chance someone finds any of this useful.

*clacks* is toy a simulation of a very simple message relaying network. Relaying 'Stations' are vertices on a graph, with edges between them representing some 'connection' media. Messages are published by Stations, who pass them on to adjacent stations, who in turn relay them on to *their* adjacent stations.


Run it
------

```
brew install zeromq protobuf
protoc --go_out=. message.proto
go run main.go
```