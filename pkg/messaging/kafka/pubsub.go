// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	kf "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"google.golang.org/protobuf/proto"
)

const (
	chansPrefix = "channels"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed to topic")
	ErrNotSubscribed     = errors.New("not subscribed")
	ErrEmptyTopic        = errors.New("empty topic")
	ErrEmptyID           = errors.New("empty id")
)

var _ messaging.PubSub = (*pubsub)(nil)

type subscription struct {
	*kf.Consumer
	cancel func() error
}
type pubsub struct {
	publisher
	url           string
	queue         string
	logger        log.Logger
	mu            sync.Mutex
	subscriptions map[string]map[string]subscription
}

// NewPubSub returns Kafka message publisher/subscriber.
func NewPubSub(url, queue string, logger log.Logger) (messaging.PubSub, error) {
	prod, err := kf.NewProducer(&kf.ConfigMap{"bootstrap.servers": url})
	if err != nil {
		return &pubsub{}, err
	}

	ret := &pubsub{
		publisher: publisher{
			prod: prod,
		},
		url:           url,
		queue:         queue,
		subscriptions: make(map[string]map[string]subscription),
		logger:        logger,
	}
	return ret, nil
}

func (ps *pubsub) Subscribe(id, topic string, handler messaging.MessageHandler) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check topic
	s, ok := ps.subscriptions[topic]
	switch ok {
	case true:
		// Check topic ID
		if _, ok := s[id]; ok {
			return ErrAlreadySubscribed
		}
	default:
		s = make(map[string]subscription)
		ps.subscriptions[topic] = s
	}

	consumer, err := kf.NewConsumer(&kf.ConfigMap{
		"bootstrap.servers":        ps.url,
		"broker.address.family":    "v4",
		"group.id":                 "mainflux",
		"auto.offset.reset":        "latest",
		"metadata.max.age.ms":      1,
		"allow.auto.create.topics": "true",
	})
	if err != nil {
		return err
	}
	if err = consumer.SubscribeTopics([]string{formatTopic(topic)}, nil); err != nil {
		return err
	}

	go func() {
		for {
			message, err := consumer.ReadMessage(-1)
			ps.handle(message, handler)
			if err == nil {
				ps.handle(message, handler)
			} else if !err.(kf.Error).IsTimeout() {
				ps.logger.Error(err.Error())
				return
			}
		}
	}()
	s[id] = subscription{
		Consumer: consumer,
		cancel:   handler.Cancel,
	}
	return nil
}

func (ps *pubsub) Unsubscribe(id, topic string) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	// Check topic
	subs, ok := ps.subscriptions[topic]
	if !ok {
		return ErrNotSubscribed
	}
	// Check topic ID
	s, ok := subs[id]
	if !ok {
		return ErrNotSubscribed
	}
	if err := s.close(); err != nil {
		return err
	}
	delete(subs, id)
	if len(subs) == 0 {
		delete(ps.subscriptions, topic)
	}
	return nil
}

func (ps *pubsub) Close() error {
	for _, subs := range ps.subscriptions {
		for _, s := range subs {
			if err := s.close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ps *pubsub) handle(message *kf.Message, h messaging.MessageHandler) {
	var msg = &messaging.Message{}
	if err := proto.Unmarshal(message.Value, msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
	}
	if err := h.Handle(msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
	}
}

func (s subscription) close() error {
	if s.cancel != nil {
		if err := s.cancel(); err != nil {
			return err
		}
	}
	return s.Close()
}

func formatTopic(topic string) string {
	if strings.Contains(topic, "*") {
		return fmt.Sprintf("^%s", topic)
	}
	return topic
}
