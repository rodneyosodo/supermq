// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	streamID                       = "mainflux.mqtt"
	streamLen                      = 1000
	checkUnpublishedEventsInterval = 1 * time.Minute
)

type EventStore interface {
	Connect(ctx context.Context, clientID string) error
	Disconnect(ctx context.Context, clientID string) error
}

// EventStore is a struct used to store event streams in Redis.
type eventStore struct {
	client            *redis.Client
	instance          string
	unpublishedEvents []*redis.XAddArgs
}

// NewEventStore returns wrapper around mProxy service that sends
// events to event store.
func NewEventStore(ctx context.Context, client *redis.Client, instance string) EventStore {
	es := &eventStore{
		client:   client,
		instance: instance,
	}

	go es.startPublishingRoutine(ctx)

	return es
}

// Connect issues event on MQTT CONNECT.
func (es *eventStore) Connect(ctx context.Context, clientID string) error {
	return es.publish(ctx, clientID, "connect")
}

// Disconnect issues event on MQTT CONNECT.
func (es *eventStore) Disconnect(ctx context.Context, clientID string) error {
	return es.publish(ctx, clientID, "disconnect")
}

func (es *eventStore) checkRedisConnection(ctx context.Context) error {
	// A timeout is used to avoid blocking the main thread
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	return es.client.Ping(ctx).Err()
}

func (es *eventStore) publish(ctx context.Context, clientID, eventType string) error {
	event := mqttEvent{
		clientID:  clientID,
		eventType: eventType,
		instance:  es.instance,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}

	if err := es.checkRedisConnection(ctx); err != nil {
		es.unpublishedEvents = append(es.unpublishedEvents, record)
		return nil
	}

	return es.client.XAdd(ctx, record).Err()
}

func (es *eventStore) startPublishingRoutine(ctx context.Context) {
	ticker := time.NewTicker(checkUnpublishedEventsInterval)
	for {
		select {
		case <-ticker.C:
			if err := es.checkRedisConnection(ctx); err == nil {
				for i := len(es.unpublishedEvents) - 1; i >= 0; i-- {
					if err := es.client.XAdd(ctx, es.unpublishedEvents[i]).Err(); err == nil {
						es.unpublishedEvents = append(es.unpublishedEvents[:i], es.unpublishedEvents[i+1:]...)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
