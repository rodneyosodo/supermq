// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"errors"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

const (
	defBrokerType = "rabbitmq"
	envBrokerType = "MF_BROKER_TYPE"
)

var (
	errEmptyBrokerType = errors.New("Empty broker type")
)

// NewPublisher This aggregates the NewPublisher function for all message brokers
func NewPublisher(url string) (nats.Publisher, error) {
	brokerselection := mainflux.Env(envBrokerType, defBrokerType)
	if brokerselection == "nats" {
		pb, err := nats.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else if brokerselection == "rabbitmq" {
		pb, err := rabbitmq.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else {
		return nil, errEmptyBrokerType
	}
}
