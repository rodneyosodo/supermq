// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/assert"
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
	pubsubcases := []struct {
		desc         string
		topic        string
		errorMessage error
		pubsub       bool //true for publish and false for subscribe
	}{
		{
			desc:         "Susbcribe to a topic",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Susbcribe to an already subscribed topic",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			errorMessage: errors.New("already subscribed to topic"),
			pubsub:       false,
		},
		{
			desc:         "Susbcribe to a topic with a subtopic",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Susbcribe to an already subscribed topic with a subtopic",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			errorMessage: errors.New("already subscribed to topic"),
			pubsub:       false,
		},
		{
			desc:         "Susbcribe to an empty topic",
			topic:        "",
			errorMessage: errors.New("empty topic"),
			pubsub:       false,
		},
		{
			desc:         "Unsubscribe to an empty topic",
			topic:        "",
			errorMessage: errors.New("empty topic"),
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to a topic",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			errorMessage: nil,
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to an already unsubscribed topic",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			errorMessage: errors.New("not subscribed"),
			pubsub:       true,
		},
		{
			desc:         "Unsubscribe to a topic with a subtopic",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			errorMessage: nil,
			pubsub:       true,
		},
		{
			desc:         "Doubling Susbcribe to a topic",
			topic:        "increaseTopic",
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Doubling Susbcribe to an already subscribed topic",
			topic:        "increaseTopic",
			errorMessage: errors.New("already subscribed to topic"),
			pubsub:       false,
		},
		{
			desc:         "Doubling Susbcribe to a topic with a subtopic",
			topic:        "secondTopic",
			errorMessage: nil,
			pubsub:       false,
		},
		{
			desc:         "Doubling Susbcribe to an already subscribed topic with a subtopic",
			topic:        "secondTopic",
			errorMessage: errors.New("already subscribed to topic"),
			pubsub:       false,
		},
		{
			desc:         "Doubling Susbcribe to an empty topic",
			topic:        "",
			errorMessage: errors.New("empty topic"),
			pubsub:       false,
		},
		{
			desc:         "Doubling Unsubscribe to an empty topic",
			topic:        "",
			errorMessage: errors.New("empty topic"),
			pubsub:       true,
		},
		{
			desc:         "Doubling Unsubscribe to a topic",
			topic:        "increaseTopic",
			errorMessage: nil,
			pubsub:       true,
		},
		{
			desc:         "Doubling Unsubscribe to an already unsubscribed topic",
			topic:        "increaseTopic",
			errorMessage: errors.New("not subscribed"),
			pubsub:       true,
		},
		{
			desc:         "Doubling Unsubscribe to a topic with a subtopic",
			topic:        "secondTopic",
			errorMessage: nil,
			pubsub:       true,
		},
	}

	cases := []struct {
		desc         string
		topic        string
		errorMessage error
		channel      string
		subtopic     string
		payload      []byte
	}{
		{
			desc:         "publish message with nil payload",
			topic:        topic,
			errorMessage: nil,
			payload:      nil,
		},
		{
			desc:         "publish message with string payload",
			topic:        topic,
			errorMessage: nil,
			payload:      data,
		},
		{
			desc:         "publish message with channel",
			topic:        topic,
			errorMessage: nil,
			payload:      data,
			channel:      channel,
		},
		{
			desc:         "publish message with subtopic",
			topic:        topic,
			errorMessage: nil,
			payload:      data,
			subtopic:     subtopic,
		},
		{
			desc:         "publish message with channel and subtopic",
			topic:        topic,
			errorMessage: nil,
			payload:      data,
			channel:      channel,
			subtopic:     subtopic,
		},
		{
			desc:         "publish message with nil topic",
			topic:        "",
			errorMessage: errors.New("empty topic"),
			payload:      nil,
		},
	}
	for _, pc := range pubsubcases {
		if pc.pubsub == false {
			err := pubsub.Subscribe(pc.topic, handler)
			if pc.errorMessage == nil {
				require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
			} else {
				assert.Equal(t, err, pc.errorMessage)
			}
		} else {
			err := pubsub.Unsubscribe(pc.topic)
			if pc.errorMessage == nil {
				require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
			} else {
				assert.Equal(t, err, pc.errorMessage)
			}
		}
	}
	for _, tc := range cases {
		expectedMsg := messaging.Message{
			Channel:  tc.channel,
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		_ = pubsub.Subscribe(tc.topic, handler)
		err := pubsub.Publish(tc.topic, expectedMsg)
		if tc.errorMessage == nil {
			require.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s", tc.desc, err))
			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		} else {
			assert.Equal(t, err, tc.errorMessage)
		}
	}
	expectedMsg := messaging.Message{
		Channel:  channel,
		Subtopic: "demo",
		Payload:  data,
	}
	err := pubsub.Unsubscribe(topic)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Publish(topic, expectedMsg)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Publish(topic, expectedMsg)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
}

func handler(msg messaging.Message) error {
	msgChan <- msg
	return nil
}
