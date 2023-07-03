// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package tracing

import (
	"context"
	"net"

	"github.com/mainflux/mainflux/internal/server"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"go.opentelemetry.io/otel/trace"
)

// Constants to define different operations to be traced.
const (
	subscribeOP   = "subscribe_op"
	unsubscribeOp = "unsubscribe_op"
	handleOp      = "handle_op"
)

var _ messaging.PubSub = (*pubsubMiddleware)(nil)

type pubsubMiddleware struct {
	publisherMiddleware
	pubsub messaging.PubSub
	host   hostConfig
}

// NewPubSub creates a new pubsub middleware that traces pubsub operations.
func NewPubSub(config server.Config, tracer trace.Tracer, pubsub messaging.PubSub) (messaging.PubSub, error) {
	pb := &pubsubMiddleware{
		publisherMiddleware: publisherMiddleware{
			publisher: pubsub,
			tracer:    tracer,
			host: hostConfig{
				host: config.Host,
				port: config.Port,
			},
		},
		pubsub: pubsub,
		host: hostConfig{
			host: config.Host,
			port: config.Port,
		},
	}

	ipAddrs, err := net.LookupIP(config.Host)
	if err != nil {
		return pb, errors.Wrap(ErrFailedToLookupIP, err)
	}

	for _, ipv4Addr := range ipAddrs {
		if ipv4Addr.To4() != nil {
			pb.host.IPV4 = ipv4Addr.String()
			pb.publisherMiddleware.host.IPV4 = ipv4Addr.String()
		}
	}
	for _, ipv6Addr := range ipAddrs {
		if ipv6Addr.To16() != nil && ipv6Addr.To4() == nil {
			pb.host.IPV6 = ipv6Addr.String()
			pb.publisherMiddleware.host.IPV6 = ipv6Addr.String()
		}
	}

	return pb, nil
}

// Subscribe creates a new subscription and traces the operation.
func (pm *pubsubMiddleware) Subscribe(ctx context.Context, id string, topic string, handler messaging.MessageHandler) error {
	ctx, span := createSpan(ctx, pm.host, subscribeOP, topic, "", id, pm.tracer)
	defer span.End()
	h := &traceHandler{
		handler: handler,
		tracer:  pm.tracer,
		ctx:     ctx,
		host:    pm.host,
	}
	return pm.pubsub.Subscribe(ctx, id, topic, h)
}

// Unsubscribe removes an existing subscription and traces the operation.
func (pm *pubsubMiddleware) Unsubscribe(ctx context.Context, id string, topic string) error {
	ctx, span := createSpan(ctx, pm.host, unsubscribeOp, topic, "", id, pm.tracer)
	defer span.End()
	return pm.pubsub.Unsubscribe(ctx, id, topic)
}

// TraceHandler is used to trace the message handling operation.
type traceHandler struct {
	handler messaging.MessageHandler
	tracer  trace.Tracer
	ctx     context.Context
	topic   string
	host    hostConfig
}

// Handle instruments the message handling operation.
func (h *traceHandler) Handle(msg *messaging.Message) error {
	_, span := createSpan(h.ctx, h.host, handleOp, h.topic, msg.Subtopic, msg.Publisher, h.tracer)
	defer span.End()
	return h.handler.Handle(msg)
}

// Cancel cancels the message handling operation.
func (h *traceHandler) Cancel() error {
	return h.handler.Cancel()
}
