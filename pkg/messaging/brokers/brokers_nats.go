//go:build nats
// +build nats

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package brokers

import (
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.>"

type (
	Publisher nats.Publisher
	PubSub    nats.PubSub
)

func NewPublisher(url string) (Publisher, error) {
	pb, err := nats.NewPublisher(url)
	if err != nil {
		return nil, err
	}
	return pb, nil

}

func NewPubSub(url, queue string, logger log.Logger) (PubSub, error) {
	pb, err := nats.NewPubSub(url, queue, logger)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
