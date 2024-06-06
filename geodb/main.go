package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
)

func main() {
	worldMap := NewMap()

	StartServer(worldMap)
}

func StartServer(worldMap *Map) {
	log.Println("Starting server on port 8080")
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	waitGroup := sync.WaitGroup{}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			HandleConnection(conn, worldMap)
		}()
	}
}

func HandleConnection(conn net.Conn, worldMap *Map) {
	log.Println("Handling connection from: ", conn.RemoteAddr())
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		HandleCommand(line, conn, worldMap)
	}
}

type Command struct {
	Command  string          `json:"command"`
	Location *LocationEntity `json:"locationEntity"`
	GridName string          `json:"gridName"`
}

// { "command": "add", "locationEntity": { "locationId": "123", "latitude": 37.7749, "longitude": -122.4194 } }
// 86283082fffffff
// { "command": "listen", "gridName": "86283082fffffff" }
func HandleCommand(line string, conn net.Conn, worldMap *Map) {
	command, err := ParseCommand(line)
	if err != nil {
		log.Println("Error parsing command: ", err, line)
		conn.Write([]byte("Error parsing command\n"))
		return
	}

	switch command.Command {
	case LocationChangeTypeAdd:
		log.Println("Adding location: ", command.Location)
		worldMap.AddLocation(command.Location)
	case ListenGridUpdate:
		log.Println("Listening for grid updates: ", command.GridName)
		var subscriber = make(chan LocationChangeEvent)
		worldMap.Subscribe(command.GridName, subscriber)
		go func() {
			for {
				update := <-subscriber
				jsonString, err := json.Marshal(update)
				if err != nil {
					log.Println("Error marshalling update: ", err)
				} else {
					conn.Write(jsonString)
				}
			}
		}()
	default:
		log.Println("Unknown command: ", command.Command)
		conn.Write([]byte("Unknown command\n"))
	}
}

func ParseCommand(line string) (Command, error) {
	//parse the command and return a Command struct
	var command Command
	err := json.Unmarshal([]byte(line), &command)
	if err != nil {
		return Command{}, err
	}

	return command, nil
}
