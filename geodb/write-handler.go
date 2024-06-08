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
	closeChan chan struct{}
}

func NewWriteHandler(world *Map) *WriteHandler {
	return &WriteHandler{
		WorldMap:  world,
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

			go handleWriteConnection(conn, worldMap)
		}
	}
}

func handleWriteConnection(conn net.Conn, worldMap *Map) {
	log.Println("New write connection from: ", conn.RemoteAddr())

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection: ", err)
		}
	}(conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		HandleWriterCommand(line, conn, worldMap)
	}
}

func HandleWriterCommand(line string, conn net.Conn, worldMap *Map) {
	var command WriteCommand
	err := json.Unmarshal([]byte(line), &command)
	if err != nil {
		log.Println("Error parsing command: ", err, line)
		_, err := conn.Write([]byte("Error parsing command\n"))
		if err != nil {
			log.Println("Error writing to connection: ", err)
			return
		}
		return
	}

	err = worldMap.Save(command.LodId, command.Lat, command.Lon)
	if err != nil {
		log.Println("Error saving location to map: ", err)
		return
	}
}
