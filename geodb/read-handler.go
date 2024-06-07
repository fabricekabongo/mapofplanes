package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
)

type SubscribeCommand struct {
	GridName string `json:"gridName"`
}

type ReadHandler struct {
	WorldMap  *Map
	closeChan chan struct{}
}

func NewReadHandler(world *Map) *ReadHandler {
	return &ReadHandler{
		WorldMap:  world,
		closeChan: make(chan struct{}),
	}
}

func (r *ReadHandler) listen(listener net.Listener, worldMap *Map) {
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
		case <-r.closeChan:
			return
		default:
			conn, err := listener.Accept()
			waitGroup.Add(1)

			if err != nil {
				panic(err)
			}

			go r.handleReadConnection(conn, worldMap)
		}
	}
}

func (r *ReadHandler) handleReadConnection(conn net.Conn, worldMap *Map) {
	log.Println("New read connection from: ", conn.RemoteAddr())

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection: ", err)
		}
	}(conn)

	scanner := bufio.NewScanner(conn) // to listen to new subscription
	writer := bufio.NewWriter(conn)   // to write events

	waitGroup := sync.WaitGroup{}
	defer waitGroup.Wait()

	addSub := make(chan LocationAddedEvent)
	updateSub := make(chan LocationUpdateEvent)
	deleteSub := make(chan LocationDeletedEvent)

	go listenToUpdate(addSub, updateSub, deleteSub, writer)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			return
		}
		var command SubscribeCommand
		err := json.Unmarshal([]byte(line), &command)

		if err != nil {
			log.Println("Error parsing command: ", err, line)
			_, err := conn.Write([]byte("Error parsing command\n"))
			if err != nil {
				return
			}

			err = conn.Close()
			if err != nil {
				return
			}
			return
		}

		grid, ok := worldMap.Grids[command.GridName] // Grid should almost always exist but okay
		if !ok {
			log.Println("Grid not found: ", command.GridName)
			_, err := conn.Write([]byte("Grid not found\n"))
			if err != nil {
				return
			}
			err = conn.Close()
			if err != nil {
				return
			}
			return
		}

		grid.addEventSubscribers[conn.RemoteAddr().String()] = addSub
		grid.updateEventSubscribers[conn.RemoteAddr().String()] = updateSub
		grid.deleteEventSubscribers[conn.RemoteAddr().String()] = deleteSub
	}
}

func listenToUpdate(addSub chan LocationAddedEvent, updateSub chan LocationUpdateEvent, deleteSub chan LocationDeletedEvent, writer *bufio.Writer) {
	for {
		select {
		case added := <-addSub:
			data, err := json.Marshal(added)
			if err != nil {
				log.Println("Error marshalling added event: ", err)
				continue
			}
			_, err = writer.Write(data)
			if err != nil {
				return
			}
		case updated := <-updateSub:
			data, err := json.Marshal(updated)
			if err != nil {
				log.Println("Error marshalling updated event: ", err)
				continue
			}
			_, err = writer.Write(data)
			if err != nil {
				return
			}
		case deleted := <-deleteSub:
			data, err := json.Marshal(deleted)
			if err != nil {
				log.Println("Error marshalling deleted event: ", err)
				continue
			}
			_, err = writer.Write(data)
			if err != nil {
				return
			}
		}
	}
}
