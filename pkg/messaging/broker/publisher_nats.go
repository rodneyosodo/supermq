//go:build nats
// +build nats

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"github.com/mainflux/mainflux/pkg/messaging/nats"
)

func NewPublisher(url string) (nats.Publisher, error) {
	pb, err := nats.NewPublisher(url)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
