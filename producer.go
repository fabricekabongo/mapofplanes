package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	address    string
	connection *amqp.Connection
}

func NewProducer(address string) *Producer {
	return &Producer{
		address: address,
	}
}

func (p *Producer) connect() error {
	connection, err := amqp.Dial(p.address)
	p.connection = connection

	return err
}

func (p *Producer) SendMessage(queue string, message ADSBMessage) {

}
