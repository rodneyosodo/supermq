// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"errors"
	"strings"

	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

var (
	errEmptyBrokerType = errors.New("Empty broker type")
)

// NewPublisher This aggregates the NewPublisher function for all message brokers
func NewPublisher(url string) (nats.Publisher, error) {
	if strings.Contains(url, "nats") {
		pb, err := nats.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else if strings.Contains(url, "rabbitmq") {
		pb, err := rabbitmq.NewPublisher(url)
		if err != nil {
			return nil, err
		}
		return pb, nil
	} else {
		return nil, errEmptyBrokerType
	}
}
