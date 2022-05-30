// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package brokers

import (
	"errors"
	"time"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/mqtt"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.>"

var errEmptyBrokerType = errors.New("empty broker type")

// Publisher interface enriched with connection closing.
type Publisher interface {
	messaging.Publisher
	Close()
}

// PubSub interface enriched with connection closing.
type PubSub interface {
	messaging.PubSub
	Close()
}

// NewPublisher This aggregates the NewPublisher function for all message brokers
func NewPublisher(brokerType, url string) (Publisher, error) {
	// brokerType := mainflux.Env(envBrokerType, defBrokerType)
	switch brokerType {
	case "nats":
		pb, err := nats.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	case "rabbitmq":
		pb, err := rabbitmq.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	case "mqtt":
		mqttTimout, err := time.ParseDuration("30s")
		if err != nil {
			return nil, err
		}
		pb, err := mqtt.NewPublisher(url, mqttTimout)
		if err != nil {
			return nil, err
		}
		return pb, nil
	default:
		return nil, errEmptyBrokerType
	}
}

// NewPubSub This aggregates the NewPubSub function for all message brokers
func NewPubSub(brokerType, url, queue string, logger log.Logger) (PubSub, error) {
	// brokerType := mainflux.Env(envBrokerType, defBrokerType)
	switch brokerType {
	case "nats":
		pb, err := nats.NewPubSub(url, queue, logger)
		if err != nil {
			return nil, err
		}
		return pb, nil
	case "rabbitmq":
		pb, err := rabbitmq.NewPubSub(url, queue, logger)
		if err != nil {
			return nil, err
		}
		return pb, nil
	case "mqtt":
		mqttTimout, err := time.ParseDuration("30s")
		if err != nil {
			return nil, err
		}
		pb, err := mqtt.NewPubSub(url, queue, mqttTimout, logger)
		if err != nil {
			return nil, err
		}
		return pb, nil
	default:
		return nil, errEmptyBrokerType
	}
}
