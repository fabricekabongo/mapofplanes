package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ADSBClient struct {
	Address    string
	Port       string
	Connection net.Conn
}

func NewADSBClient(address string, port string) *ADSBClient {
	return &ADSBClient{
		Address: address,
		Port:    port,
	}
}

func (client *ADSBClient) Connect() error {
	dialer := &net.Dialer{
		KeepAlive: 5 * time.Minute, // Set your desired keepalive time
	}
	connection, err := dialer.Dial("tcp", client.Address+":"+client.Port)

	if err != nil {
		log.Println("failed to connect to server", err)
		return err
	}

	client.Connection = connection

	return nil
}

func (client *ADSBClient) StartListening(workers int) {
	reader := bufio.NewReader(client.Connection)
	waitGroup := sync.WaitGroup{}
	workChannel := make(chan string, workers)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("failed to read message", err)
			continue
		}
		workChannel <- message
		waitGroup.Add(1)

		go func(message string) {
			defer waitGroup.Done()
			defer func() { <-workChannel }()

			parsedMessage, err := client.parseMessage(message)
			if err != nil {
				panic(err)
			}

			fmt.Println(parsedMessage)
		}(message)
	}
}

func (client *ADSBClient) parseMessage(message string) (ADSBMessage, error) {
	var messageParts = strings.Split(message, ",")

	if len(messageParts) != 22 {
		return ADSBMessage{}, fmt.Errorf("invalid message format")
	}

	transmissionType, err := strconv.Atoi(messageParts[1])
	if err != nil {
		return ADSBMessage{}, err
	}

	altitude, _ := strconv.Atoi(messageParts[11])
	groundSpeed, _ := strconv.Atoi(messageParts[12])
	track, _ := strconv.Atoi(messageParts[13])
	latitude, _ := strconv.Atoi(messageParts[14])
	longitude, _ := strconv.Atoi(messageParts[15])
	verticalRate, _ := strconv.Atoi(messageParts[16])

	alert, _ := strconv.ParseBool(messageParts[18])
	emergency, _ := strconv.ParseBool(messageParts[19])
	spi, _ := strconv.ParseBool(messageParts[20])
	isOnGround, _ := strconv.ParseBool(messageParts[21])

	var adsbMessage = ADSBMessage{
		MessageType:          MESSAGE_TYPE(messageParts[0]),
		TransmissionType:     TRANSMISSION_TYPE(transmissionType),
		SessionId:            messageParts[2],
		AircraftId:           messageParts[3],
		HexIdent:             messageParts[4],
		FlightId:             messageParts[5],
		DateMessageGenerated: messageParts[6],
		TimeMessageGenerated: messageParts[7],
		DateMessageLogged:    messageParts[8],
		TimeMessageLogged:    messageParts[9],
		CallSign:             messageParts[10],
		Altitude:             altitude,
		GroundSpeed:          groundSpeed,
		Track:                track,
		Latitude:             latitude,
		Longitude:            longitude,
		VerticalRate:         verticalRate,
		Squawk:               messageParts[17],
		Alert:                alert,
		Emergency:            emergency,
		Spi:                  spi,
		IsOnGround:           isOnGround,
	}

	return adsbMessage, nil
}
