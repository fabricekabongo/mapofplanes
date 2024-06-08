package server

import (
	"log"
	"net"
)

type Server struct {
	WriteHandler *WriteHandler
	ReadHandler  *ReadHandler
	closeChannel chan struct{}
}

func NewServer(wHandler WriteHandler, rHander ReadHandler) *Server {
	return &Server{
		WriteHandler: &wHandler,
		ReadHandler:  &rHander,
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

	go s.WriteHandler.listen(writerListener)
	go s.ReadHandler.listen(subscriberListener)

	<-s.closeChannel
}
