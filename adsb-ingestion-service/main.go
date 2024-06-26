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
	rabbitmqUrl   = os.Getenv("RABBITMQ_URL")
	rabbitmqQueue = os.Getenv("RABBITMQ_QUEUE")
	GeoDBUrl      = os.Getenv("GEODB_URL")
	RedisUrl      = os.Getenv("REDIS_URL")
)

func main() {
	// check if all environment variables are set
	// get the missing environment variables from flags
	populateEnv()

	consumer := NewConsumer(rabbitmqUrl, rabbitmqQueue)

	err := consumer.Connect()
	if err != nil {
		panic(err)
	}
	processor := NewSBS1Processor(GeoDBUrl, RedisUrl, consumer.MessagesChannel)
	err = processor.Connect()
	if err != nil {
		panic(err)
	}

	defer prepareTermination(consumer, processor)

	log.Println("Connected to RabbitMQ server")
	log.Println("Starting to listen for messages")
	go consumer.StartListening(10)

	log.Println("Starting to process messages")
	go processor.Start()

	// listen for sigterm
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down...")
}

func populateEnv() {
	if rabbitmqUrl == "" || rabbitmqQueue == "" || GeoDBUrl == "" || RedisUrl == "" {
		log.Println("Please set the following environment variables:")
		log.Println("RABBITMQ_URL, RABBITMQ_QUEUE, GEODB_URL, REDIS_URL")
		log.Println("Reverting to flags...")

		flag.StringVar(&rabbitmqUrl, "rabbitmq-url", "amqp://user:userpassword@localhost:5672/", "RabbitMQ URL")
		flag.StringVar(&rabbitmqQueue, "rabbitmq-queue", "sbs1", "RabbitMQ queue")
		flag.StringVar(&GeoDBUrl, "geodb-url", "", "GEODB URL")
		flag.StringVar(&RedisUrl, "redis-url", "", "Redis URL")
		flag.Parse()

		if rabbitmqUrl == "" || rabbitmqQueue == "" || GeoDBUrl == "" || RedisUrl == "" {
			log.Println("No flags set. Please set the following flags:")
			log.Println("rabbitmq-url, rabbitmq-queue, geodb-url, redis-url")
			os.Exit(1)
		}
	}
}

func prepareTermination(consumer *Consumer, processor *SBS1Processor) {
	log.Println("Closing connection to TCP server")
	err := consumer.Close()
	if err != nil {
		log.Println("Error closing connection to RabbitMQ server", err)
	}
	time.Sleep(3 * time.Second)
	err = processor.Close()
	if err != nil {
		log.Println("Error closing connection to Redis server", err)
	}
	log.Println("Connection closed")
}
