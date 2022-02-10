// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq_test

import (
	"fmt"
	"testing"

	// "github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/require"
)

const (
	topic        = "topic"
	chansPrefix  = "channels"
	channel      = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic     = "engine"
	routingKey   = "routinngkey"
	exchange     = "mainflux"
	exchangeKind = "fanout"
)

var (
	msgChan = make(chan messaging.Message)
	data    = []byte("payload")
)

func TestPubsub(t *testing.T) {
	err := pubsub.Subscribe(fmt.Sprintf("%s.%s", chansPrefix, topic), handler)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Subscribe(fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic), handler)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	err = pubsub.Unsubscribe(fmt.Sprintf("%s.%s", chansPrefix, topic))
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Unsubscribe(fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic))
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc     string
		channel  string
		subtopic string
		payload  []byte
	}{
		{
			desc:    "publish message with nil payload",
			payload: nil,
		},
		{
			desc:    "publish message with string payload",
			payload: data,
		},
		{
			desc:    "publish message with channel",
			payload: data,
			channel: channel,
		},
		{
			desc:     "publish message with subtopic",
			payload:  data,
			subtopic: subtopic,
		},
		{
			desc:     "publish message with channel and subtopic",
			payload:  data,
			channel:  channel,
			subtopic: subtopic,
		},
	}

	for _, tc := range cases {
		expectedMsg := messaging.Message{
			Channel:  tc.channel,
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		err = pubsub.Publish(topic, expectedMsg)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		// err = pubsub.ReadMessages(topic)
		// require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		// msgs, err := pubsub.channel.Consume(pubsub.Queue.Name, fmt.Sprintf("%s.%s", chansPrefix, topic), true, false, false, false, nil)
		// receivedMsg := <-msgChan
		// assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
}

func handler(msg messaging.Message) error {
	msgChan <- msg
	return nil
}
