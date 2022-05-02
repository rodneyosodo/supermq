// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
)

var (
	errSubscribeTimeout   = errors.New("failed to subscribe due to timeout reached")
	errUnsubscribeTimeout = errors.New("failed to unsubscribe due to timeout reached")
	errAlreadySubscribed  = errors.New("already subscribed to topic")
	errNotSubscribed      = errors.New("not subscribed")
	errEmptyTopic         = errors.New("empty topic")
	errEmptyID            = errors.New("empty ID")
)

var _ messaging.Subscriber = (*subscriber)(nil)

type subscriber struct {
	address       string
	timeout       time.Duration
	logger        log.Logger
	subscriptions map[string]map[string]mqtt.Client
}

// NewSubscriber returns a new MQTT message subscriber.
func NewSubscriber(address string, timeout time.Duration, logger log.Logger) (messaging.Subscriber, error) {
	ret := subscriber{
		address:       address,
		timeout:       timeout,
		logger:        logger,
		subscriptions: make(map[string]map[string]mqtt.Client),
	}
	return ret, nil
}

func (sub subscriber) Subscribe(id, topic string, handler messaging.MessageHandler) error {
	if id == "" {
		return errEmptyID
	}
	if topic == "" {
		return errEmptyTopic
	}
	// Check topic
	s, ok := sub.subscriptions[topic]
	if ok {
		// Check client ID
		if _, ok := s[id]; ok {
			return errAlreadySubscribed
		}
	} else {
		opts := mqtt.NewClientOptions().SetUsername(username).AddBroker(sub.address)
		client := mqtt.NewClient(opts)
		token := client.Connect()
		if token.Error() != nil {
			return token.Error()
		}
		s[id] = client
		sub.subscriptions[topic] = s
	}
	client := sub.subscriptions[topic][id]
	token := client.Subscribe(topic, qos, sub.mqttHandler(handler))
	if token.Error() != nil {
		return token.Error()
	}
	if ok := token.WaitTimeout(sub.timeout); !ok {
		return errSubscribeTimeout
	}
	return token.Error()
}

func (sub subscriber) Unsubscribe(id, topic string) error {
	if id == "" {
		return errEmptyID
	}
	if topic == "" {
		return errEmptyTopic
	}
	// Check topic
	s, ok := sub.subscriptions[topic]
	if ok {
		// Check topic ID
		_, ok := s[id]
		if !ok {
			return errNotSubscribed
		}
	} else {
		return errNotSubscribed
	}
	client := sub.subscriptions[topic][id]
	token := client.Unsubscribe(topic)
	if token.Error() != nil {
		return token.Error()
	}

	ok = token.WaitTimeout(sub.timeout)
	if !ok {
		return errUnsubscribeTimeout
	}
	delete(sub.subscriptions, id)
	return token.Error()
}

func (sub subscriber) mqttHandler(h messaging.MessageHandler) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		var msg messaging.Message
		if err := proto.Unmarshal(m.Payload(), &msg); err != nil {
			sub.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h.Handle(msg); err != nil {
			sub.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
	}
}
