package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type SBS1Processor struct {
	redis        *redis.Client
	redisURL     string
	msgChannel   chan ADSBMessage
	ctx          context.Context
	processing   bool
	closeChannel chan struct{}
}

func NewSBS1Processor(redisURL string, msgChannel chan ADSBMessage) *SBS1Processor {
	var ctx = context.Background()

	return &SBS1Processor{
		redisURL:   redisURL,
		msgChannel: msgChannel,
		ctx:        ctx,
		processing: false,
	}
}

func (p *SBS1Processor) Connect() error {
	options, err := redis.ParseURL(p.redisURL)
	if err != nil {
		return err
	}

	redisClient := redis.NewClient(options)
	p.redis = redisClient

	return nil
}

func (client *SBS1Processor) Close() error {
	client.closeChannel <- struct{}{}
	time.Sleep(3 * time.Second)
	return client.redis.Close()
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
			p.handleLocationMessage(message)
			p.handleIdentityMessage(message)
			p.handleVelocityMessage(message)
			p.handleAltitudeMessage(message)
		}
	}
}

func (p *SBS1Processor) handleLocationMessage(message ADSBMessage) error {
	if message.TransmissionType != TranmissionTypeSurfacePosition && message.TransmissionType != TranmissionTypeAirbornePosition {
		return errors.New("invalid transmission type")
	}

	if message.Latitude < -90 || message.Latitude > 90 || message.Longitude < -180 || message.Longitude > 180 {
		return errors.New("invalid location")
	}

	p.redis.JSONSet(p.ctx, message.HexIdent, ".lat", message.Latitude)
	p.redis.JSONSet(p.ctx, message.HexIdent, ".lon", message.Longitude)

	// Add location to Redis geo set
	p.redis.GeoAdd(p.ctx, "locations", &redis.GeoLocation{
		Name:      message.HexIdent,
		Latitude:  float64(message.Latitude),
		Longitude: float64(message.Longitude),
	})

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
