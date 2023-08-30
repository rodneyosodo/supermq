// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

var _ policies.Cache = (*pcache)(nil)

type pcache struct {
	client      *redis.Client
	keyDuration time.Duration
}

// NewCache returns redis policy cache implementation.
func NewCache(client *redis.Client, duration time.Duration) policies.Cache {
	return pcache{
		client:      client,
		keyDuration: duration,
	}
}

func (pc pcache) Put(ctx context.Context, key, value string) error {
	if err := pc.client.Set(ctx, key, value, pc.keyDuration).Err(); err != nil {
		return errors.Wrap(errors.ErrCreateEntity, err)
	}

	return nil
}

func (pc pcache) Get(ctx context.Context, key string) (string, error) {
	res := pc.client.Get(ctx, key)
	// Nil response indicates non-existent key in Redis client.
	if res == nil || res.Err() == redis.Nil {
		return "", errors.ErrNotFound
	}
	if err := res.Err(); err != nil {
		return "", err
	}

	return res.Result()
}

func (pc pcache) Remove(ctx context.Context, key string) error {
	if err := pc.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(errors.ErrRemoveEntity, err)
	}

	return nil
}
