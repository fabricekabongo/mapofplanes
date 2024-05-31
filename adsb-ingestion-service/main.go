package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	rabbitmqUrl   = os.Getenv("RABBITMQ_URL")
	rabbitmqQueue = os.Getenv("RABBITMQ_QUEUE")
)

func main() {
	// check if all environment variables are set
	if rabbitmqUrl == "" || rabbitmqQueue == "" {
		log.Println("Please set the following environment variables:")
		log.Println("RABBITMQ_URL, RABBITMQ_QUEUE")
		log.Println("Reverting to flags...")
		// get the missing environment variables from flags
		flag.StringVar(&rabbitmqUrl, "rabbitmq-url", "", "RabbitMQ URL")
		flag.StringVar(&rabbitmqQueue, "rabbitmq-queue", "", "RabbitMQ Queue")
		flag.Parse()

		if rabbitmqUrl == "" || rabbitmqQueue == "" {
			log.Println("No flags set. Please set the following flags:")
			log.Println("rabbitmq-url, rabbitmq-queue")
			os.Exit(1)
		}
	}

	consumer := NewConsumer(rabbitmqUrl, rabbitmqQueue)

	err := consumer.Connect()
	if err != nil {
		panic(err)
	}
	defer prepareTermination(consumer)

	log.Println("Connected to RabbitMQ server")
	log.Println("Starting to listen for messages")
	go consumer.StartListening(10)

	log.Println("Starting to process messages")

	go func() {
		for {
			message := <-consumer.MessagesChannel
			log.Println("Received message", message)
		}
	}()

	// listen for sigterm
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
}

func prepareTermination(consumer *Consumer) {
	log.Println("Closing connection to TCP server")
	consumer.Close()
	log.Println("Connection closed")
}
