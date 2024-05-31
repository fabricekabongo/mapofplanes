package main

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	address    string
	queue      string
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewProducer(address string, queue string) *Producer {
	return &Producer{
		address: address,
		queue:   queue,
	}
}

func (p *Producer) Close() {
	p.connection.Close()
}

func (p *Producer) connect() error {
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

func (p *Producer) SendMessage(message ADSBMessage) {

	body, err := json.Marshal(message)

	if err != nil {
		panic(err)
	}

	err = p.channel.Publish("", p.queue, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
		Expiration:  "5000", // 5 seconds
	})

	if err != nil {
		panic(err)
	}
}
