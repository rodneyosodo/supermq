// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	checkUnpublishedEventsInterval = 1 * time.Minute
	checkRedisConnectionInterval   = 100 * time.Millisecond
)

// Event represents redis event.
type Event interface {
	// Encode encodes event to map.
	Encode() (map[string]interface{}, error)
}

// Publisher specifies redis event publishing API.
type Publisher interface {
	// Publishes event to redis stream.
	Publish(ctx context.Context, event Event) error

	// StartPublishingRoutine starts routine that checks for unpublished events
	// and publishes them to redis stream.
	StartPublishingRoutine(ctx context.Context)
}

type eventStore struct {
	client            *redis.Client
	unpublishedEvents []*redis.XAddArgs
	streamID          string
	streamLen         int64
	mu                sync.Mutex
}

func NewEventStore(client *redis.Client, streamID string, streamLen int64) Publisher {
	return &eventStore{
		client:    client,
		streamID:  streamID,
		streamLen: streamLen,
	}
}

func (es *eventStore) Publish(ctx context.Context, event Event) error {
	values, err := event.Encode()
	if err != nil {
		return err
	}
	record := &redis.XAddArgs{
		Stream:       es.streamID,
		MaxLenApprox: es.streamLen,
		Values:       values,
	}

	if err := es.checkRedisConnection(ctx); err != nil {
		es.unpublishedEvents = append(es.unpublishedEvents, record)
		return nil
	}

	return es.client.XAdd(ctx, record).Err()
}

func (es *eventStore) StartPublishingRoutine(ctx context.Context) {
	ticker := time.NewTicker(checkUnpublishedEventsInterval)
	for {
		select {
		case <-ticker.C:
			if err := es.checkRedisConnection(ctx); err == nil {
				es.mu.Lock()
				for i := len(es.unpublishedEvents) - 1; i >= 0; i-- {
					if err := es.client.XAdd(ctx, es.unpublishedEvents[i]).Err(); err == nil {
						es.unpublishedEvents = append(es.unpublishedEvents[:i], es.unpublishedEvents[i+1:]...)
					}
				}
				es.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}
func (es *eventStore) checkRedisConnection(ctx context.Context) error {
	// A timeout is used to avoid blocking the main thread
	ctx, cancel := context.WithTimeout(ctx, checkRedisConnectionInterval)
	defer cancel()

	return es.client.Ping(ctx).Err()
}
