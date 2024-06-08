package main

import (
	"log"
	"net"
)

type Server struct {
	WorldMap     *Map
	closeChannel chan struct{}
	Cluster      *Cluster
}

func NewServer(world *Map, cluster Cluster) *Server {
	return &Server{
		WorldMap:     world,
		Cluster:      &cluster,
		closeChannel: make(chan struct{}),
	}
}

func (s *Server) Stop() {
	s.closeChannel <- struct{}{}
	close(s.closeChannel)
}

func (s *Server) Start() {
	log.Println("Starting database server on port 19999 for write and 20000 for read")

	writerListener, err := net.Listen("tcp", ":19999")
	if err != nil {
		panic(err)
	}

	subscriberListener, err := net.Listen("tcp", ":20000")
	if err != nil {
		panic(err)
	}

	WriteHandler := NewWriteHandler(s.WorldMap, s.Cluster)
	ReadHandler := NewReadHandler(s.WorldMap)

	go WriteHandler.listen(writerListener, s.WorldMap)
	go ReadHandler.listen(subscriberListener, s.WorldMap)

	<-s.closeChannel
}
