package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"time"
)

var (
	FailedToWriteToGeoDB       = errors.New("failed to write to GeoDB")
	FailedToWriteToRedis       = errors.New("failed to write to Redis")
	InvalidLocationCoordinates = errors.New("invalid location coordinates")
)

type SBS1Processor struct {
	geoDB        net.Conn
	geoDBUrl     string
	redis        redis.Client
	redisUrl     string
	msgChannel   chan ADSBMessage
	ctx          context.Context
	processing   bool
	closeChannel chan struct{}
}

func NewSBS1Processor(geoDBUrl string, redisUrl string, msgChannel chan ADSBMessage) *SBS1Processor {
	var ctx = context.Background()

	return &SBS1Processor{
		geoDBUrl:   geoDBUrl,
		redisUrl:   redisUrl,
		msgChannel: msgChannel,
		ctx:        ctx,
		processing: false,
	}
}

func (p *SBS1Processor) Connect() error {
	err := p.connectToGeoDB()
	if err != nil {
		return err
	}

	err = p.connectToRedis()
	if err != nil {
		return err
	}

	return nil
}

func (p *SBS1Processor) connectToGeoDB() error {
	geoDB, err := net.Dial("tcp", p.geoDBUrl)
	if err != nil {
		return err
	}
	p.geoDB = geoDB

	return nil
}

func (p *SBS1Processor) connectToRedis() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: p.redisUrl,
	})

	_, err := redisClient.Ping(p.ctx).Result()
	if err != nil {
		return err
	}

	p.redis = *redisClient

	return nil
}

func (client *SBS1Processor) Close() error {
	client.closeChannel <- struct{}{}
	time.Sleep(3 * time.Second)
	err := client.geoDB.Close()
	if err != nil {
		log.Println("failed to close connection to TCP server", err)
	}
	err = client.redis.Close()
	if err != nil {
		log.Println("failed to close connection to Redis server", err)
	}

	return err
}

func (p *SBS1Processor) Start() {
	p.processing = true

	go func() {
		<-p.closeChannel
		p.processing = false
	}()

	for p.processing {
		message := <-p.msgChannel
		storedMessage, err := p.redis.JSONGet(p.ctx, message.HexIdent, "$").Result()
		if err != nil {
			continue
		}

		if storedMessage == "" {
			p.redis.JSONSet(p.ctx, message.HexIdent, "$", message)
			continue
		} else {
			log.Println("Message found in Redis. Updating:")
			p.redis.JSONSet(p.ctx, message.HexIdent, ".generatedDate", message.DateMessageGenerated)
			p.redis.JSONSet(p.ctx, message.HexIdent, ".generatedTime", message.TimeMessageGenerated)

			err = p.handleLocationMessage(message)
			if err != nil {
				log.Println("Failed to handle location message", err)
				if errors.Is(err, FailedToWriteToGeoDB) {
					time.Sleep(10 * time.Second)
					err := p.connectToGeoDB()
					if err != nil {
						log.Fatalln("Failed to reconnect to GeoDB", err)
					}
				}
			}
			_ = p.handleIdentityMessage(message)
			_ = p.handleVelocityMessage(message)
			_ = p.handleAltitudeMessage(message)
		}
	}
}

func (p *SBS1Processor) handleLocationMessage(message ADSBMessage) error {
	writer := bufio.NewWriter(p.geoDB)

	if message.TransmissionType != TranmissionTypeSurfacePosition && message.TransmissionType != TranmissionTypeAirbornePosition {
		return nil
	}

	if message.Latitude < -90 || message.Latitude > 90 || message.Longitude < -180 || message.Longitude > 180 {
		return InvalidLocationCoordinates
	}

	write, err := writer.Write([]byte(fmt.Sprintf("{\"loc_id\":\"%s\",\"lat\":%f,\"lon\":%f}\n", message.HexIdent, message.Latitude, message.Longitude)))
	if err != nil {
		log.Println("Failed to write to GeoDB", err)
		return FailedToWriteToGeoDB
	}

	err = writer.Flush()

	if err != nil {
		log.Println("Failed to flush to GeoDB", err)
		return FailedToWriteToGeoDB
	}

	log.Println("Wrote to GeoDB", write)

	return nil
}

func (p *SBS1Processor) handleIdentityMessage(message ADSBMessage) error {
	if message.TransmissionType != TransmissionTypeIdentityAndCategory {
		return errors.New("invalid transmission type")
	}

	p.redis.JSONSet(p.ctx, message.HexIdent, ".callsign", message.CallSign)

	return nil
}

func (p *SBS1Processor) handleVelocityMessage(message ADSBMessage) error {
	if message.TransmissionType != TranmissionTypeAirborneVelocity {
		return errors.New("invalid transmission type")
	}

	p.redis.JSONSet(p.ctx, message.HexIdent, ".groundSpeed", message.GroundSpeed)
	p.redis.JSONSet(p.ctx, message.HexIdent, ".track", message.Track)
	p.redis.JSONSet(p.ctx, message.HexIdent, ".verticalRate", message.VerticalRate)

	return nil
}

func (p *SBS1Processor) handleAltitudeMessage(message ADSBMessage) error {
	if message.TransmissionType != TranmissionTypeSurveillanceAltitude {
		return errors.New("invalid transmission type")
	}

	p.redis.JSONSet(p.ctx, message.HexIdent, ".altitude", message.Altitude)

	return nil
}
