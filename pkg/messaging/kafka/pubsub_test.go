// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic       = "topic"
	chansPrefix = "channels"
	channel     = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic    = "engine"
	clientID    = "819b273a-b8213f9eda3b"
)

var (
	msgChan = make(chan *messaging.Message)
	data    = []byte("payload")
)

func TestPubsub(t *testing.T) {
	err := pubsub.Subscribe(context.TODO(), clientID, fmt.Sprintf("%s.*", chansPrefix), handler{})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc     string
		topic    string
		subtopic string
		payload  []byte
		err      error
	}{
		{
			desc:    "publish message with empty topic",
			topic:   "",
			err:     kafka.ErrEmptyTopic,
			payload: data,
		},
		{
			desc:    "publish message with nil payload",
			topic:   channel,
			payload: nil,
			err:     nil,
		},
		{
			desc:    "publish message with string payload",
			topic:   channel,
			payload: data,
			err:     nil,
		},
		{
			desc:    "publish message with topic",
			payload: data,
			topic:   channel,
			err:     nil,
		},
		{
			desc:     "publish message with subtopic",
			payload:  data,
			subtopic: subtopic,
			err:      kafka.ErrEmptyTopic,
		},
		{
			desc:     "publish message with topic and subtopic",
			payload:  data,
			topic:    channel,
			subtopic: subtopic,
			err:      nil,
		},
	}

	for _, tc := range cases {
		expectedMsg := messaging.Message{
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		err := publisher.Publish(context.TODO(), tc.topic, &expectedMsg)
		if tc.err == nil {
			require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", tc.desc, err))
		} else {
			assert.Equal(t, err, tc.err)
		}
	}

	// Test Subscribe and Unsubscribe
	subcases := []struct {
		desc         string
		topic        string
		topicID      string
		errorMessage error
		pubsub       bool // true for subscribe and false for unsubscribe
	}{
		{
			desc:         "Subscribe to a topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid1",
			errorMessage: nil,
			pubsub:       true,
		},
		{
			desc:         "Subscribe to the same topic with a different ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid2",
			errorMessage: nil,
			pubsub:       true,
		},
		{
			desc:         "Subscribe to an already subscribed topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid1",
			errorMessage: kafka.ErrAlreadySubscribed,
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to a topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid1",
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Unsubscribe to a non-existent topic with an ID",
			topic:        "h",
			topicID:      "topicid1",
			errorMessage: kafka.ErrNotSubscribed,
			pubsub:       false,
		},
		{
			desc:         "Unsubscribe to the same topic with a different ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid2",
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Unsubscribe to the same topic with a different ID not subscribed",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid3",
			errorMessage: kafka.ErrNotSubscribed,
			pubsub:       false,
		},
		{
			desc:         "Unsubscribe to an already unsubscribed topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "topicid1",
			errorMessage: kafka.ErrNotSubscribed,
			pubsub:       false,
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			topicID:      "topicid1",
			errorMessage: nil,
			pubsub:       true,
		},
		{
			desc:         "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			topicID:      "topicid1",
			errorMessage: kafka.ErrAlreadySubscribed,
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			topicID:      "topicid1",
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Unsubscribe to an already unsubscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			topicID:      "topicid1",
			errorMessage: kafka.ErrNotSubscribed,
			pubsub:       false,
		},
		{
			desc:         "Subscribe to an empty topic with an ID",
			topic:        "",
			topicID:      "topicid1",
			errorMessage: kafka.ErrEmptyTopic,
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to an empty topic with an ID",
			topic:        "",
			topicID:      "topicid1",
			errorMessage: kafka.ErrEmptyTopic,
			pubsub:       false,
		},
		{
			desc:         "Subscribe to a topic with empty id",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "",
			errorMessage: kafka.ErrEmptyID,
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to a topic with empty id",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			topicID:      "",
			errorMessage: kafka.ErrEmptyID,
			pubsub:       false,
		},
	}

	for _, pc := range subcases {
		if pc.pubsub == true {
			err := pubsub.Subscribe(context.TODO(), pc.topicID, pc.topic, handler{})
			if pc.errorMessage == nil {
				require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
			} else {
				assert.Equal(t, err, pc.errorMessage)
			}
		} else {
			err := pubsub.Unsubscribe(context.TODO(), pc.topicID, pc.topic)
			if pc.errorMessage == nil {
				require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
			} else {
				assert.Equal(t, err, pc.errorMessage)
			}
		}
	}
}

type handler struct{}

func (h handler) Handle(msg *messaging.Message) error {
	fmt.Printf("Message %s\n", msg.String())
	msgChan <- msg
	return nil
}

func (h handler) Cancel() error {
	return nil
}
