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
	Address         string
	Port            string
	Connection      net.Conn
	MessagesChannel chan ADSBMessage
	closeChannel    chan struct{}
}

func NewADSBClient(address string, port string) *ADSBClient {
	return &ADSBClient{
		Address:         address,
		Port:            port,
		MessagesChannel: make(chan ADSBMessage, 2000),
		closeChannel:    make(chan struct{}, 1),
	}
}

func (client *ADSBClient) Close() error {
	client.closeChannel <- struct{}{}
	time.Sleep(3 * time.Second)
	return client.Connection.Close()
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
	continueScheduling := true
	failureCount := 0

	go func() {
		<-client.closeChannel
		continueScheduling = false
	}()

	for continueScheduling {
		message, err := reader.ReadString('\n')
		if err != nil {
			if failureCount > 10 {
				panic(err)
			}

			failureCount++
			log.Println("failed to read message", err)
			log.Println("Will wait 10 seconds before trying again")
			time.Sleep(10 * time.Second)
			log.Println("Trying to reconnect")
			client.Connection.Close()
			err := client.Connect()
			if err != nil {
				log.Println("failed to reconnect", err)
				log.Println("Will wait 10 seconds before trying again")
				time.Sleep(10 * time.Second)
				continue
			}

			log.Println("failed to read message", err)
			continue
		}

		failureCount = 0
		workChannel <- message
		waitGroup.Add(1)

		go func(message string) {
			defer waitGroup.Done()
			defer func() { <-workChannel }()

			parsedMessage, err := client.parseMessage(message)
			if err != nil {
				panic(err)
			}

			client.MessagesChannel <- parsedMessage
		}(message)
	}

	waitGroup.Wait()
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

	altitude, _ := strconv.ParseFloat(messageParts[11], 64)
	groundSpeed, _ := strconv.ParseFloat(messageParts[12], 64)
	track, _ := strconv.Atoi(messageParts[13])
	latitude, _ := strconv.ParseFloat(messageParts[14], 64)
	longitude, _ := strconv.ParseFloat(messageParts[15], 64)
	verticalRate, _ := strconv.ParseFloat(messageParts[16], 64)

	alert, _ := strconv.ParseBool(messageParts[18])
	emergency, _ := strconv.ParseBool(messageParts[19])
	spi, _ := strconv.ParseBool(messageParts[20])
	isOnGround, _ := strconv.ParseBool(messageParts[21])

	var adsbMessage = ADSBMessage{
		MessageType:          messageParts[0],
		TransmissionType:     transmissionType,
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
