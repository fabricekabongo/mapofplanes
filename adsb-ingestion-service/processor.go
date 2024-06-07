package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type SBS1Processor struct {
	geoDB        net.Conn
	geoDBUrl     string
	msgChannel   chan ADSBMessage
	ctx          context.Context
	processing   bool
	closeChannel chan struct{}
}

func NewSBS1Processor(geoDBUrl string, msgChannel chan ADSBMessage) *SBS1Processor {
	var ctx = context.Background()

	return &SBS1Processor{
		geoDBUrl:   geoDBUrl,
		msgChannel: msgChannel,
		ctx:        ctx,
		processing: false,
	}
}

func (p *SBS1Processor) Connect() error {
	geoDB, err := net.Dial("tcp", p.geoDBUrl)
	if err != nil {
		return err
	}
	p.geoDB = geoDB

	return nil
}

func (client *SBS1Processor) Close() error {
	client.closeChannel <- struct{}{}
	time.Sleep(3 * time.Second)
	return client.geoDB.Close()
}

func (p *SBS1Processor) Start() {
	p.processing = true

	go func() {
		<-p.closeChannel
		p.processing = false
	}()

	for p.processing {
		message := <-p.msgChannel
		p.handleLocationMessage(message)
		//storedMessage, err := p.geoDB.JSONGet(p.ctx, message.HexIdent, "$").Result()
		//if err != nil {
		//	continue
		//}
		//
		//if storedMessage == "" {
		//	p.geoDB.JSONSet(p.ctx, message.HexIdent, "$", message)
		//	continue
		//} else {
		//	log.Println("Message found in Redis. Updating:")
		//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".generatedDate", message.DateMessageGenerated)
		//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".generatedTime", message.TimeMessageGenerated)
		//	p.handleLocationMessage(message)
		//	p.handleIdentityMessage(message)
		//	p.handleVelocityMessage(message)
		//	p.handleAltitudeMessage(message)
		//}
	}
}

func (p *SBS1Processor) handleLocationMessage(message ADSBMessage) error {
	writer := bufio.NewWriter(p.geoDB)

	if message.TransmissionType != TranmissionTypeSurfacePosition && message.TransmissionType != TranmissionTypeAirbornePosition {
		return errors.New("invalid transmission type")
	}

	if message.Latitude < -90 || message.Latitude > 90 || message.Longitude < -180 || message.Longitude > 180 {
		return errors.New("invalid location")
	}

	write, err := writer.Write([]byte(fmt.Sprintf("{\"loc_id\":\"%s\",\"lat\":%f,\"lon\":%f}\n", message.HexIdent, message.Latitude, message.Longitude)))
	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	log.Println("Wrote to GeoDB", write)

	return nil
}

//func (p *SBS1Processor) handleIdentityMessage(message ADSBMessage) error {
//	if message.TransmissionType != TransmissionTypeIdentityAndCategory {
//		return errors.New("invalid transmission type")
//	}
//
//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".callsign", message.CallSign)
//
//	return nil
//}

//func (p *SBS1Processor) handleVelocityMessage(message ADSBMessage) error {
//	if message.TransmissionType != TranmissionTypeAirborneVelocity {
//		return errors.New("invalid transmission type")
//	}
//
//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".groundSpeed", message.GroundSpeed)
//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".track", message.Track)
//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".verticalRate", message.VerticalRate)
//
//	return nil
//}

//func (p *SBS1Processor) handleAltitudeMessage(message ADSBMessage) error {
//	if message.TransmissionType != TranmissionTypeSurveillanceAltitude {
//		return errors.New("invalid transmission type")
//	}
//
//	p.geoDB.JSONSet(p.ctx, message.HexIdent, ".altitude", message.Altitude)
//
//	return nil
//}
