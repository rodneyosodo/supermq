//go:build rabbitmq
// +build rabbitmq

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

func NewPublisher(url string) (rabbitmq.Publisher, error) {
	pb, err := rabbitmq.NewPublisher(url)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
