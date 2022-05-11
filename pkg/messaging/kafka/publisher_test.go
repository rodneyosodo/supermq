// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher(t *testing.T) {
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
			errorMessage: errors.New("empty topic"),
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
		err := publisher.Publish(tc.topic, expectedMsg)
		if tc.errorMessage == nil {
			require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", tc.desc, err))
			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		} else {
			assert.Equal(t, err, tc.errorMessage)
		}
	}
}
