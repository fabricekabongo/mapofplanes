package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	adsbHost      = os.Getenv("ADSB_HOST")
	adsbPort      = os.Getenv("ADSB_PORT")
	rabbitmqUrl   = os.Getenv("RABBITMQ_URL")
	rabbitmqQueue = os.Getenv("RABBITMQ_QUEUE")
)

func main() {
	// check if all environment variables are set
	if adsbHost == "" || adsbPort == "" || rabbitmqUrl == "" || rabbitmqQueue == "" {
		log.Println("Please set the following environment variables:")
		log.Println("ADSB_HOST, ADSB_PORT, RABBITMQ_URL, RABBITMQ_QUEUE")
		log.Println("Reverting to flags")
		flag.StringVar(&adsbHost, "adsb-host", "localhost", "ADSB host")
		flag.StringVar(&adsbPort, "adsb-port", "30003", "ADSB port")
		flag.StringVar(&rabbitmqUrl, "rabbitmq-url", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL")
		flag.StringVar(&rabbitmqQueue, "rabbitmq-queue", "adsb", "RabbitMQ queue")
		flag.Parse()

		if adsbHost == "" || adsbPort == "" || rabbitmqUrl == "" || rabbitmqQueue == "" {
			log.Println("Please provide all flags")
			log.Println("Usage: adsb-producer --adsbHost=localhost --adsbPort=30003 --rabbitmqUrl=amqp://guest:guest@localhost:5672/ --rabbitmqQueue=adsb")
			os.Exit(1)
		}
	}

	log.Println("started adsb producer...")
	client := NewADSBClient(adsbHost, adsbPort)
	err := client.Connect()
	if err != nil {
		panic(err)
	}

	producer := NewProducer(rabbitmqUrl, rabbitmqQueue)

	err = producer.connect()
	if err != nil {
		panic(err)
	}
	defer prepareTermination(client, producer)

	log.Println("Connected to TCP server")
	log.Println("Starting to listen for messages...")
	go client.StartListening(10)

	log.Println("Starting to send messages...")

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
	log.Println("Shutting down...")
}

func prepareTermination(client *ADSBClient, producer *Producer) {
	log.Println("Closing connection to TCP server")
	client.Close()

	// make sure there are note messages waiting to be sent before closing the connection
	log.Println("Waiting for messages to be sent...")
	time.Sleep(3 * time.Second)
	log.Println("Closing connection to RabbitMQ server")
	producer.Close()
	log.Println("Connection closed")
}
