package main

import (
	"encoding/json"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	address         string
	queue           string
	connection      *amqp.Connection
	channel         *amqp.Channel
	MessagesChannel chan ADSBMessage
	closeChannel    chan struct{}
}

func NewConsumer(address string, queue string) *Consumer {
	return &Consumer{
		address:         address,
		queue:           queue,
		MessagesChannel: make(chan ADSBMessage, 50),
		closeChannel:    make(chan struct{}, 1),
	}
}

func (p *Consumer) Close() error {
	return p.connection.Close()
}

func (p *Consumer) Connect() error {
	connection, err := amqp.Dial(p.address)
	if err != nil {
		return err
	}

	p.connection = connection

	channel, err := p.connection.Channel()
	if err != nil {
		return err
	}
	args := make(amqp.Table)

	args["x-message-ttl"] = 5000            // 5 seconds
	args["x-max-length"] = 500000           // 500k messages
	args["x-overflow"] = "drop-head"        // drop oldest message
	args["x-max-length-bytes"] = 2000000000 // 2GB

	_, err = channel.QueueDeclare(p.queue, true, false, false, false, args)
	if err != nil {
		return err
	}

	p.channel = channel

	return nil
}

func (p *Consumer) StartListening(workers int) error {
	waitGroup := sync.WaitGroup{}
	workChannel := make(chan amqp.Delivery, workers)
	continueScheduling := true
	failureCount := 0

	go func() {
		<-p.closeChannel
		continueScheduling = false
	}()

	delivery, err := p.channel.Consume(p.queue, "adsb-ingestion-service", false, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	for continueScheduling {
		d := <-delivery
		workChannel <- d

		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			defer func() { <-workChannel }()

			var message ADSBMessage

			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				if d.Redelivered {
					err = d.Nack(false, false)
					if err != nil {
						failureCount++
						if failureCount > 10 {
							panic(err)
						}

						log.Println("failed to NACK message", err)
					} else {
						failureCount = 0
					}
				} else {
					err := d.Nack(false, true)
					if err != nil {
						failureCount++
						if failureCount > 10 {
							panic(err)
						}

						log.Println("failed to NACK message", err)
					} else {
						failureCount = 0
					}
				}
			}

			p.MessagesChannel <- message
			err = d.Ack(false)
			if err != nil {
				failureCount++
				if failureCount > 10 {
					panic(err)
				}

				log.Println("failed to ACK message", err)
			} else {
				failureCount = 0
			}
		}()
	}

	waitGroup.Wait()

	return nil
}
