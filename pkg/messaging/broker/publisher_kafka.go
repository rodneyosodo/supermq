//go:build kafka
// +build kafka

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"github.com/mainflux/mainflux/pkg/messaging/kafka"
)

func NewPublisher(url string) (kafka.Publisher, error) {
	pb, err := kafka.NewPublisher(url)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
