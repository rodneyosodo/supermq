// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"errors"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/messaging/kafka"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

const (
	envBrokerType = "MF_BROKER_TYPE"
)

var (
	errEmptyBrokerType = errors.New("empty broker type")
	defBrokerType      = "nats"
)

// NewPublisher This aggregates the NewPublisher function for all message brokers
func NewPublisher(url string) (nats.Publisher, error) {
	brokerSelection := mainflux.Env(envBrokerType, defBrokerType)
	if brokerSelection == "nats" {
		pb, err := nats.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else if brokerSelection == "rabbitmq" {
		pb, err := rabbitmq.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else if brokerSelection == "kafka" {
		pb, err := kafka.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else {
		return nil, errEmptyBrokerType
	}
}
