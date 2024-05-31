package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	adsbHost         = os.Getenv("ADSB_HOST")
	adsbPort         = os.Getenv("ADSB_PORT")
	rabbitmqHost     = os.Getenv("RABBITMQ_HOST")
	rabbitmqPort     = os.Getenv("RABBITMQ_PORT")
	rabbitmqUser     = os.Getenv("RABBITMQ_USER")
	rabbitmqPassword = os.Getenv("RABBITMQ_PASSWORD")
	rabbitmqQueue    = os.Getenv("RABBITMQ_QUEUE")
)

func main() {

	fmt.Println("started adsb producer...")
	client := NewADSBClient(adsbHost, adsbPort)
	err := client.Connect()
	if err != nil {
		panic(err)
	}
	rabbitConnectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqUser, rabbitmqPassword, rabbitmqHost, rabbitmqPort)
	producer := NewProducer(rabbitConnectionString, rabbitmqQueue)

	err = producer.connect()
	if err != nil {
		panic(err)
	}
	defer prepareTermination(client, producer)

	fmt.Println("Connected to TCP server")
	fmt.Println("Starting to listen for messages...")
	go client.StartListening(10)

	fmt.Println("Starting to send messages...")

	go func() {
		for {
			message := <-client.MessagesChannel
			producer.SendMessage(message)
		}
	}()

	// listen for sigterm
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Shutting down...")
}

func prepareTermination(client *ADSBClient, producer *Producer) {
	fmt.Println("Closing connection to TCP server")
	client.Close()

	// make sure there are note messages waiting to be sent before closing the connection
	fmt.Println("Waiting for messages to be sent...")
	time.Sleep(3 * time.Second)
	fmt.Println("Closing connection to RabbitMQ server")
	producer.Close()
	fmt.Println("Connection closed")
}
