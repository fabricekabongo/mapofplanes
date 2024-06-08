package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
)

type WriteCommand struct {
	LodId string  `json:"loc_id"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
}

type WriteHandler struct {
	WorldMap  *Map
	Cluster   *Cluster
	closeChan chan struct{}
}

func NewWriteHandler(world *Map, cluster *Cluster) *WriteHandler {
	return &WriteHandler{
		WorldMap:  world,
		Cluster:   cluster,
		closeChan: make(chan struct{}),
	}
}

func (w *WriteHandler) listen(listener net.Listener, worldMap *Map) {
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Println("Error closing listener: ", err)
		}
	}(listener)

	go w.Cluster.Start()
	go func() {
		for {
			select {
			case msg := <-w.Cluster.Broadcast:
				w.handleWriterCommand(msg)
			}
		}
	}()
	waitGroup := sync.WaitGroup{}

	defer waitGroup.Wait()

	for {
		select {
		case <-w.closeChan:
			return
		default:
			conn, err := listener.Accept()
			waitGroup.Add(1)

			if err != nil {
				panic(err)
			}

			go w.handleWriteConnection(conn)
		}
	}
}

func (w *WriteHandler) handleWriteConnection(conn net.Conn) {
	log.Println("New write connection from: ", conn.RemoteAddr())

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection: ", err)
		}
	}(conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			break
		}

		w.handleWriterCommand(line)
	}
}

func (w *WriteHandler) handleWriterCommand(line []byte) {
	go w.Cluster.BroadcastMessage(line)

	var command WriteCommand
	err := json.Unmarshal(line, &command)

	if err != nil {
		log.Println("Error parsing command: ", err, line)
		return
	}

	err = w.WorldMap.Save(command.LodId, command.Lat, command.Lon)
	if err != nil {
		log.Println("Error saving location to map: ", err)
		return
	}
}
