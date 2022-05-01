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
	client        mqtt.Client
	timeout       time.Duration
	logger        log.Logger
	subscriptions map[string]map[string]mqtt.Token
}

// NewSubscriber returns a new MQTT message subscriber.
func NewSubscriber(address string, timeout time.Duration, logger log.Logger) (messaging.Subscriber, error) {
	client, err := newClient(address, timeout)
	if err != nil {
		return nil, err
	}

	ret := subscriber{
		client:        client,
		timeout:       timeout,
		logger:        logger,
		subscriptions: make(map[string]map[string]mqtt.Token),
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
	s, ok := sub.subscriptions[topic]
	if !ok {
		s = make(map[string]mqtt.Token)
	}
	if _, ok := s[id]; ok {
		return errAlreadySubscribed
	}
	token := sub.client.Subscribe(topic, qos, sub.mqttHandler(handler))
	if token.Error() != nil {
		return token.Error()
	}
	ok = token.WaitTimeout(sub.timeout)
	if !ok {
		return errSubscribeTimeout
	}
	s[id] = token
	sub.subscriptions[topic] = s
	return token.Error()
}

func (sub subscriber) Unsubscribe(id, topic string) error {
	if id == "" {
		return errEmptyID
	}
	if topic == "" {
		return errEmptyTopic
	}
	s, ok := sub.subscriptions[topic]
	if !ok {
		return errNotSubscribed
	}
	if _, ok := s[id]; !ok {
		return errNotSubscribed
	}
	token := sub.client.Unsubscribe(topic)
	if token.Error() != nil {
		return token.Error()
	}

	ok = token.WaitTimeout(sub.timeout)
	if !ok {
		return errUnsubscribeTimeout
	}
	delete(s, id)
	if len(s) == 0 {
		delete(sub.subscriptions, topic)
	} else {
		sub.subscriptions[topic] = s
	}
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
