// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package tracing

import (
	"context"
	"fmt"

	"github.com/mainflux/mainflux/internal/server"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Traced operations.
const publishOP = "publish"

var defaultAttributes = []attribute.KeyValue{
	attribute.String("messaging.system", "nats"),
	attribute.Bool("messaging.destination.anonymous", false),
	attribute.String("messaging.destination.template", "channels/{channelID}/messages/*"),
	attribute.Bool("messaging.destination.temporary", true),
	attribute.String("network.protocol.name", "nats"),
	attribute.String("network.protocol.version", "2.2.4"),
	attribute.String("network.transport", "tcp"),
	attribute.String("network.type", "ipv4"),
}

// ErrFailedToLookupIP is returned when the IP address of the database peer cannot be looked up.
var ErrFailedToLookupIP = errors.New("failed to lookup IP address")

var _ messaging.Publisher = (*publisherMiddleware)(nil)

type publisherMiddleware struct {
	publisher messaging.Publisher
	tracer    trace.Tracer
	host      server.Config
}

// New creates new messaging publisher tracing middleware.
func New(config server.Config, tracer trace.Tracer, publisher messaging.Publisher) messaging.Publisher {
	pub := &publisherMiddleware{
		publisher: publisher,
		tracer:    tracer,
		host:      config,
	}

	return pub
}

// Publish traces NATS publish operations.
func (pm *publisherMiddleware) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	ctx, span := createSpan(ctx, publishOP, msg.Publisher, topic, msg.Subtopic, len(msg.Payload), pm.host, trace.SpanKindClient, pm.tracer)
	defer span.End()

	return pm.publisher.Publish(ctx, topic, msg)
}

// Close NATS trace publisher middleware.
func (pm *publisherMiddleware) Close() error {
	return pm.publisher.Close()
}

func createSpan(ctx context.Context, operation, clientID, topic, subTopic string, msgSize int, cfg server.Config, spanKind trace.SpanKind, tracer trace.Tracer) (context.Context, trace.Span) {
	var subject = fmt.Sprintf("channels.%s.messages", topic)
	if subTopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subTopic)
	}
	spanName := fmt.Sprintf("%s %s", subject, operation)

	kvOpts := []attribute.KeyValue{
		attribute.String("messaging.operation", operation),
		attribute.String("messaging.client_id", clientID),
		attribute.String("messaging.destination.name", subject),
		attribute.String("server.address", cfg.Host),
		attribute.String("server.socket.port", cfg.Port),
	}

	if msgSize > 0 {
		kvOpts = append(kvOpts, attribute.Int("messaging.message.payload_size_bytes", msgSize))
	}

	kvOpts = append(kvOpts, defaultAttributes...)

	return tracer.Start(ctx, spanName, trace.WithAttributes(kvOpts...), trace.WithSpanKind(spanKind))
}
