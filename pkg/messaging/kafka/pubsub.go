// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	log "github.com/mainflux/mainflux/logger"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	kafka "github.com/segmentio/kafka-go"
)

const (
	chansPrefix = "channels"
	offset      = kafka.LastOffset
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed to topic")
	ErrNotSubscribed     = errors.New("not subscribed")
	ErrEmptyTopic        = errors.New("empty topic")
	ErrEmptyID           = errors.New("empty id")
)

var _ messaging.PubSub = (*pubsub)(nil)

type subscription struct {
	*kafka.Reader
	cancel func() error
}
type pubsub struct {
	publisher
	logger        log.Logger
	mu            sync.Mutex
	subscriptions map[string]map[string]subscription
}

// NewPubSub returns Kafka message publisher/subscriber.
func NewPubSub(url, queue string, logger log.Logger) (messaging.PubSub, error) {
	conn, err := kafka.Dial("tcp", url)
	if err != nil {
		return &pubsub{}, err
	}
	ret := &pubsub{
		publisher: publisher{
			url:    url,
			conn:   conn,
			topics: make(map[string]*kafka.Writer),
		},
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

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{ps.url},
		GroupID:     id,
		Topic:       topic,
		StartOffset: offset,
	})
	go func() {
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				break
			}
			ps.handle(message, handler)
		}
	}()
	s[id] = subscription{
		Reader: reader,
		cancel: handler.Cancel,
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

func (ps *pubsub) handle(message kafka.Message, h messaging.MessageHandler) {
	var msg messaging.Message
	if err := proto.Unmarshal(message.Value, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
	}
	if err := h.Handle(&msg); err != nil {
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
