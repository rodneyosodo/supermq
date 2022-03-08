// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"

	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

const (
	// chansPrefix = "channels"

	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.>"
)

// PubSub type
type PubSub nats.PubSub

// NewPubSub This aggregates the NewPubSub function for all message brokers
func NewPubSub(url, queue string, logger log.Logger) (nats.PubSub, error) {
	brokerSelection := mainflux.Env(envBrokerType, defBrokerType)
	if brokerSelection == "nats" {
		pb, err := nats.NewPubSub(url, queue, logger)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else if brokerSelection == "rabbitmq" {
		pb, err := rabbitmq.NewPubSub(url, queue, logger)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else {
		return nil, errEmptyBrokerType
	}
}
