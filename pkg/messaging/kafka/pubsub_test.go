// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/require"
)

const (
	topic       = "topic"
	chansPrefix = "channels"
	channel     = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic    = "engine"
)

var (
	msgChan = make(chan messaging.Message)
	data    = []byte("payload")
)

func TestPubsub(t *testing.T) {
	expectedMsg := messaging.Message{
		Channel:  channel,
		Subtopic: "demo",
		Payload:  data,
	}
	err := pubsub.Subscribe(topic, handler)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Publish(topic, expectedMsg)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Publish(topic, expectedMsg)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Unsubscribe(topic)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
}

func handler(msg messaging.Message) error {
	msgChan <- msg
	return nil
}
