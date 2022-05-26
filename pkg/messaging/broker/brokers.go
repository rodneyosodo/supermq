// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"errors"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

const (
	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.>"
)

var (
	errEmptyBrokerType = errors.New("empty broker type")
)

// PubSub type
type PubSub nats.PubSub

// NewPublisher This aggregates the NewPublisher function for all message brokers
func NewPublisher(brokerType, url string) (nats.Publisher, error) {
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
	default:
		return nil, errEmptyBrokerType
	}
}

// NewPubSub This aggregates the NewPubSub function for all message brokers
func NewPubSub(brokerType, url, queue string, logger log.Logger) (nats.PubSub, error) {
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
	default:
		return nil, errEmptyBrokerType
	}
}
