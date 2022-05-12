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
	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.>"
	partition          = 0
	groupID            = "mainflux"
	offset             = kafka.LastOffset
)

var (
	errAlreadySubscribed = errors.New("already subscribed to topic")
	errNotSubscribed     = errors.New("not subscribed")
	errEmptyTopic        = errors.New("empty topic")
	errEmptyID           = errors.New("empty ID")
)

var _ messaging.PubSub = (*pubsub)(nil)

// PubSub wraps messaging Publisher exposing
// Close() method for Kafka connection.
type PubSub interface {
	messaging.PubSub
	Close()
}

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
func NewPubSub(url, queue string, logger log.Logger) (PubSub, error) {
	conn, err := kafka.Dial("tcp", url)
	if err != nil {
		return nil, err
	}
	ret := &pubsub{
		publisher: publisher{
			conn: conn,
			url:  url,
		},
		subscriptions: make(map[string]map[string]subscription),
		logger:        logger,
	}
	return ret, nil
}

func (ps *pubsub) Subscribe(id, topic string, handler messaging.MessageHandler) error {
	if id == "" {
		return errEmptyID
	}
	if topic == "" {
		return errEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check topic
	s, ok := ps.subscriptions[topic]
	switch ok {
	case true:
		// Check topic ID
		if _, ok := s[id]; ok {
			return errAlreadySubscribed
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
		return errEmptyID
	}
	if topic == "" {
		return errEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	// Check topic
	s, ok := ps.subscriptions[topic]
	if !ok {
		return errNotSubscribed
	}
	// Check topic ID
	reader, ok := s[id]
	if !ok {
		return errNotSubscribed
	}
	if reader.cancel != nil {
		if err := reader.cancel(); err != nil {
			return err
		}
	}
	if err := reader.Close(); err != nil {
		return err
	}
	delete(s, id)
	if len(s) == 0 {
		delete(ps.subscriptions, topic)
	}
	return nil
}

func (ps *pubsub) handle(message kafka.Message, h messaging.MessageHandler) {
	var msg messaging.Message
	if err := proto.Unmarshal(message.Value, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
	}
	if err := h.Handle(msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
	}
}
