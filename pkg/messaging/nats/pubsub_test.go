// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic       = "topic"
	chansPrefix = "channels"
	channel     = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic    = "engine"
	clientID    = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
)

var (
	msgChan   = make(chan messaging.Message)
	data      = []byte("payload")
	errFailed = errors.New("failed")
)

func TestPublisher(t *testing.T) {
	err := pubsub.Subscribe(clientID, fmt.Sprintf("%s.%s", chansPrefix, topic), handler{})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Subscribe(clientID, fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic), handler{})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc         string
		topic        string
		channel      string
		subtopic     string
		payload      []byte
		errorMessage error
	}{
		{
			desc:         "publish message with nil topic",
			topic:        "",
			payload:      nil,
			errorMessage: nats.ErrEmptyTopic,
		},
		{
			desc:         "publish message with nil payload",
			topic:        topic,
			payload:      nil,
			errorMessage: nil,
		},
		{
			desc:         "publish message with string payload",
			topic:        topic,
			payload:      data,
			errorMessage: nil,
		},
		{
			desc:         "publish message with channel",
			topic:        topic,
			payload:      data,
			channel:      channel,
			errorMessage: nil,
		},
		{
			desc:         "publish message with subtopic",
			topic:        topic,
			payload:      data,
			subtopic:     subtopic,
			errorMessage: nil,
		},
		{
			desc:         "publish message with channel and subtopic",
			topic:        topic,
			payload:      data,
			channel:      channel,
			subtopic:     subtopic,
			errorMessage: nil,
		},
	}

	for _, tc := range cases {
		expectedMsg := messaging.Message{
			Channel:  tc.channel,
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		if tc.errorMessage == nil {
			err = pubsub.Publish(tc.topic, expectedMsg)
			require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		} else {
			err = pubsub.Publish(tc.topic, expectedMsg)
			assert.Equal(t, err, tc.errorMessage)
		}
	}
}

func TestPubsub(t *testing.T) {
	// Test Subscribe and Unsubscribe
	subcases := []struct {
		desc         string
		topic        string
		clientID     string
		errorMessage error
		pubsub       bool //true for subscribe and false for unsubscribe
		handler      messaging.MessageHandler
	}{
		{
			desc:         "Subscribe to a topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to the same topic with a different ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid2",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an already subscribed topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a non-existent topic with an ID",
			topic:        "h",
			clientID:     "clientid1",
			errorMessage: nats.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientidd2",
			errorMessage: nats.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID not subscribed",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientidd3",
			errorMessage: nats.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nats.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientidd1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientidd1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientidd1",
			errorMessage: nil,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientid1",
			errorMessage: nats.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an empty topic with an ID",
			topic:        "",
			clientID:     "clientid1",
			errorMessage: nats.ErrEmptyTopic,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an empty topic with an ID",
			topic:        "",
			clientID:     "clientid1",
			errorMessage: nats.ErrEmptyTopic,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to a topic with empty id",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "",
			errorMessage: nats.ErrEmptyID,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with empty id",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "",
			errorMessage: nats.ErrEmptyID,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to another topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"1"),
			clientID:     "clientid3",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to another already subscribed topic with an ID with Unsubscribe failing",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"1"),
			clientID:     "clientid3",
			errorMessage: errFailed,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to a new topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"2"),
			clientID:     "clientid4",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Unsubscribe from a topic with an ID with failing handler",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"2"),
			clientID:     "clientid4",
			errorMessage: errFailed,
			pubsub:       false,
			handler:      handler{true},
		},
	}

	for _, pc := range subcases {
		if pc.pubsub == true {
			err := pubsub.Subscribe(pc.clientID, pc.topic, pc.handler)
			if pc.errorMessage == nil {
				require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
			} else {
				assert.Equal(t, err, pc.errorMessage)
			}
		} else {
			err := pubsub.Unsubscribe(pc.clientID, pc.topic)
			if pc.errorMessage == nil {
				require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
			} else {
				assert.Equal(t, err, pc.errorMessage)
			}
		}
	}
}

type handler struct {
	fail bool
}

func (h handler) Handle(msg messaging.Message) error {
	msgChan <- msg
	return nil
}

func (h handler) Cancel() error {
	if h.fail {
		return errFailed
	}
	return nil
}
