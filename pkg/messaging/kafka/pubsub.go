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
)

var _ messaging.PubSub = (*pubsub)(nil)

// PubSub wraps messaging Publisher exposing
// Close() method for Kafka connection.
type PubSub interface {
	messaging.PubSub
	Close()
}

type pubsub struct {
	publisher
	mu            sync.Mutex
	subscriptions map[string]*kafka.Reader
	logger        log.Logger
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
			writer: &kafka.Writer{
				Addr:  kafka.TCP(url),
				Async: true,
			},
		},
		subscriptions: make(map[string]*kafka.Reader),
		logger:        logger,
	}
	return ret, nil
}

func (ps *pubsub) Subscribe(topic string, handler messaging.MessageHandler) error {
	if topic == "" {
		return errEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, ok := ps.subscriptions[topic]; ok {
		return errAlreadySubscribed
	}
	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	groupID := fmt.Sprintf("%s.%s", groupID, topic)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{ps.publisher.url},
		GroupID:     groupID,
		Topic:       subject,
		StartOffset: offset,
	})
	ps.subscriptions[topic] = reader
	go func() {
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				break
			}
			ps.handle(message, handler)
		}
	}()
	return nil
}

func (ps *pubsub) Unsubscribe(topic string) error {
	if topic == "" {
		return errEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	reader, ok := ps.subscriptions[topic]
	if !ok {
		return errNotSubscribed
	}
	if err := reader.Close(); err != nil {
		return err
	}
	delete(ps.subscriptions, topic)
	return nil
}

func (ps *pubsub) Close() {
	ps.conn.Close()
}

func (ps *pubsub) handle(message kafka.Message, h messaging.MessageHandler) {
	var msg messaging.Message
	if err := proto.Unmarshal(message.Value, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
		return
	}
	if err := h(msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
	}
}
