// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/absmach/supermq/pkg/messaging"
)

type MockPubSub struct{}

// NewPubSub returns mock message PubSub.
func NewPubSub() messaging.PubSub {
	return MockPubSub{}
}

func (pub MockPubSub) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	return nil
}

func (pub MockPubSub) Subscribe(ctx context.Context, cfg messaging.SubscriberConfig) error {
	return nil
}

func (pub MockPubSub) Unsubscribe(ctx context.Context, id, topic string) error {
	return nil
}

func (pub MockPubSub) Close() error {
	return nil
}
