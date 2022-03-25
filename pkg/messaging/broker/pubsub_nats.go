//go:build nats
// +build nats

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
)

const (
	// chansPrefix = "channels"

	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.>"
)

type PubSub nats.PubSub

func NewPubSub(url, queue string, logger log.Logger) (PubSub, error) {
	pb, err := nats.NewPubSub(url, queue, logger)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
