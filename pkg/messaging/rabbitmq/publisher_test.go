// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq_test

import (
	"fmt"
	"log"
	"testing"

	// "github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func TestPublisher(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("rabbitmq", "3.9.10", []string{})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}
	handleInterrupt(pool, container)

	address := fmt.Sprintf("%s:%s", "localhost", container.GetPort("5672/tcp"))
	if err := pool.Retry(func() error {
		publisher, err = rabbitmq.NewPublisher(address)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

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
		err := publisher.Publish(topic, expectedMsg)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		// msgs, err := publisher.channel.Consume(pubsub.Queue.Name, fmt.Sprintf("%s.%s", chansPrefix, topic), true, false, false, false, nil)
		// msgChan := <-msgs
		// receivedMsg := <-msgChan
		// assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
}
