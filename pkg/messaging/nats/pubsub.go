// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	broker "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

const chansPrefix = "channels"

// Publisher and Subscriber errors.
var (
	ErrNotSubscribed = errors.New("not subscribed")
	ErrEmptyTopic    = errors.New("empty topic")
	ErrEmptyID       = errors.New("empty id")

	jsStreamConfig = jetstream.StreamConfig{
		Name:              "channels",
		Description:       "Mainflux stream for sending and receiving messages in between Mainflux channels",
		Subjects:          []string{"channels.>"},
		Retention:         jetstream.LimitsPolicy,
		MaxMsgsPerSubject: 1e6,
		MaxAge:            time.Hour * 24,
		MaxMsgSize:        1024 * 1024,
		Discard:           jetstream.DiscardOld,
		Storage:           jetstream.FileStorage,
	}
)

var _ messaging.PubSub = (*pubsub)(nil)

type subscription struct {
	jetstream.ConsumeContext
	cancel func() error
}

type pubsub struct {
	publisher
	logger        mflog.Logger
	mu            sync.Mutex
	stream        jetstream.Stream
	subscriptions map[string]map[string]subscription
}

// NewPubSub returns NATS message publisher/subscriber.
// Parameter queue specifies the queue for the Subscribe method.
// If queue is specified (is not an empty string), Subscribe method
// will execute NATS QueueSubscribe which is conceptually different
// from ordinary subscribe. For more information, please take a look
// here: https://docs.nats.io/developing-with-nats/receiving/queues.
// If the queue is empty, Subscribe will be used.
func NewPubSub(ctx context.Context, url string, logger mflog.Logger) (messaging.PubSub, error) {
	conn, err := broker.Connect(url, broker.MaxReconnects(maxReconnects))
	if err != nil {
		return nil, err
	}
	js, err := jetstream.New(conn)
	if err != nil {
		return nil, err
	}
	stream, err := js.CreateStream(ctx, jsStreamConfig)
	if err != nil {
		return nil, err
	}

	ret := &pubsub{
		publisher: publisher{
			js: js,
		},
		stream:        stream,
		logger:        logger,
		subscriptions: make(map[string]map[string]subscription),
	}

	return ret, nil
}

func (ps *pubsub) Subscribe(ctx context.Context, id, topic string, handler messaging.MessageHandler) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}

	ps.mu.Lock()
	// Check topic
	s, ok := ps.subscriptions[topic]
	if ok {
		// Check client ID
		if _, ok := s[id]; ok {
			// Unlocking, so that Unsubscribe() can access ps.subscriptions
			ps.mu.Unlock()
			if err := ps.Unsubscribe(ctx, id, topic); err != nil {
				return err
			}

			ps.mu.Lock()
			// value of s can be changed while ps.mu is unlocked
			s = ps.subscriptions[topic]
		}
	}
	defer ps.mu.Unlock()
	if s == nil {
		s = make(map[string]subscription)
		ps.subscriptions[topic] = s
	}

	nh := ps.natsHandler(handler)

	consumerConfig := jetstream.ConsumerConfig{
		Name:          id,
		Description:   fmt.Sprintf("Mainflux consumer of id %s for topic %s", id, topic),
		DeliverPolicy: jetstream.DeliverNewPolicy,
		FilterSubject: topic,
	}

	consumer, err := ps.stream.CreateOrUpdateConsumer(ctx, consumerConfig)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	cc, err := consumer.Consume(nh)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	s[id] = subscription{
		ConsumeContext: cc,
		cancel:         handler.Cancel,
	}

	return nil
}

func (ps *pubsub) Unsubscribe(_ context.Context, id, topic string) error {
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
	if !ok {
		return ErrNotSubscribed
	}
	// Check topic ID
	current, ok := s[id]
	if !ok {
		return ErrNotSubscribed
	}
	if current.cancel != nil {
		if err := current.cancel(); err != nil {
			return err
		}
	}
	current.Stop()

	delete(s, id)
	if len(s) == 0 {
		delete(ps.subscriptions, topic)
	}

	return nil
}

func (ps *pubsub) natsHandler(h messaging.MessageHandler) func(m jetstream.Msg) {
	return func(m jetstream.Msg) {
		var msg messaging.Message
		if err := proto.Unmarshal(m.Data(), &msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))

			return
		}

		if err := h.Handle(&msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
		if err := m.Ack(); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to ack message: %s", err))
		}
	}
}
