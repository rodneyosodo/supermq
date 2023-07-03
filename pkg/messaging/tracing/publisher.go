// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package tracing

import (
	"context"
	"net"

	"github.com/mainflux/mainflux/internal/server"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Traced operations.
const publishOP = "publish_op"

// ErrFailedToLookupIP is returned when the IP address of the database peer cannot be looked up.
var ErrFailedToLookupIP = errors.New("failed to lookup IP address")

var _ messaging.Publisher = (*publisherMiddleware)(nil)

type hostConfig struct {
	host string
	port string
	IPV4 string
	IPV6 string
}

type publisherMiddleware struct {
	publisher messaging.Publisher
	tracer    trace.Tracer
	host      hostConfig
}

// New creates new messaging publisher tracing middleware.
func New(config server.Config, tracer trace.Tracer, publisher messaging.Publisher) (messaging.Publisher, error) {
	pub := &publisherMiddleware{
		publisher: publisher,
		tracer:    tracer,
		host: hostConfig{
			host: config.Host,
			port: config.Port,
		},
	}

	ipAddrs, err := net.LookupIP(config.Host)
	if err != nil {
		return pub, errors.Wrap(ErrFailedToLookupIP, err)
	}

	for _, ipv4Addr := range ipAddrs {
		if ipv4Addr.To4() != nil {
			pub.host.IPV4 = ipv4Addr.String()
		}
	}
	for _, ipv6Addr := range ipAddrs {
		if ipv6Addr.To16() != nil && ipv6Addr.To4() == nil {
			pub.host.IPV6 = ipv6Addr.String()
		}
	}

	return pub, nil
}

// Publish traces NATS publish operations.
func (pm *publisherMiddleware) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	ctx, span := createSpan(ctx, pm.host, publishOP, topic, msg.Subtopic, msg.Publisher, pm.tracer)
	defer span.End()
	return pm.publisher.Publish(ctx, topic, msg)
}

// Close NATS trace publisher middleware.
func (pm *publisherMiddleware) Close() error {
	return pm.publisher.Close()
}

func createSpan(ctx context.Context, host hostConfig, operation, topic, subTopic, thingID string, tracer trace.Tracer) (context.Context, trace.Span) {
	kvOpts := []attribute.KeyValue{
		attribute.String("peer.address", host.host+":"+host.port),
		attribute.String("peer.hostname", host.host),
		attribute.String("peer.ipv4", host.IPV4),
		attribute.String("peer.ipv6", host.IPV6),
		attribute.String("peer.port", host.port),
		attribute.String("peer.service", "message_broker"),
	}
	switch operation {
	case publishOP:
		kvOpts = append(kvOpts, attribute.String("publisher", thingID), attribute.String("span.kind", "producer"))
	default:
		kvOpts = append(kvOpts, attribute.String("subscriber", thingID), attribute.String("span.kind", "consumer"))
	}
	kvOpts = append(kvOpts, attribute.String("topic", topic))
	if subTopic != "" {
		kvOpts = append(kvOpts, attribute.String("subtopic", topic))
	}
	return tracer.Start(ctx, operation, trace.WithAttributes(kvOpts...))
}
