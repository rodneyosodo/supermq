// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !rabbitmq
// +build !rabbitmq

package brokers

import (
	"log"

	"github.com/mainflux/mainflux/internal/server"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/messaging/nats/tracing"
	"go.opentelemetry.io/otel/trace"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.>"

func init() {
	log.Println("The binary was build using Nats as the message broker")
}

func NewPublisher(cfg server.Config, tracer trace.Tracer, url string) (messaging.Publisher, error) {
	pb, err := nats.NewPublisher(url)
	if err != nil {
		return nil, err
	}

	pb = tracing.NewPublisher(cfg, tracer, pb)

	return pb, nil

}

func NewPubSub(cfg server.Config, tracer trace.Tracer, url, queue string, logger mflog.Logger) (messaging.PubSub, error) {
	pb, err := nats.NewPubSub(url, queue, logger)
	if err != nil {
		return nil, err
	}

	pb = tracing.NewPubSub(cfg, tracer, pb)

	return pb, nil
}
