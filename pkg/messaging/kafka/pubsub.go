// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	log "github.com/mainflux/mainflux/logger"

	"github.com/mainflux/mainflux/pkg/messaging"
	kafka "github.com/segmentio/kafka-go"
	ktopics "github.com/segmentio/kafka-go/topics"
	"google.golang.org/protobuf/proto"
)

const (
	chansPrefix               = "channels"
	SubjectAllChannels        = "channels.*"
	offset                    = kafka.LastOffset
	defaultScanningIntervalMS = 500
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
	client        *kafka.Client
	logger        log.Logger
	mu            sync.Mutex
	subscriptions map[string]map[string]subscription
}

// NewPubSub returns Kafka message publisher/subscriber.
func NewPubSub(url, _ string, logger log.Logger) (messaging.PubSub, error) {
	conn, err := kafka.Dial("tcp", url)
	if err != nil {
		return &pubsub{}, err
	}
	client := &kafka.Client{
		Addr: conn.LocalAddr(),
	}
	ret := &pubsub{
		publisher: publisher{
			url:    url,
			conn:   conn,
			topics: make(map[string]*kafka.Writer),
		},
		client:        client,
		subscriptions: make(map[string]map[string]subscription),
		logger:        logger,
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
	defer ps.mu.Unlock()

	s, err := ps.checkTopic(topic, id, ErrAlreadySubscribed)
	if err != nil {
		return err
	}
	ps.configReader(id, topic, s, handler)

	// Subscribe to all topic by prediocially scanning for all topics and consuming them
	if topic == SubjectAllChannels {
		go func() {
			for {
				topics, _ := ps.listTopic()
				for _, t := range topics {
					s, err := ps.checkTopic(t, id, ErrAlreadySubscribed)
					if err == nil {
						ps.configReader(id, t, s, handler)
					}
				}
				time.Sleep(defaultScanningIntervalMS * time.Millisecond)
			}
		}()
	}
	return nil
}

func (ps *pubsub) Unsubscribe(ctx context.Context, id, topic string) error {
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

func (ps *pubsub) listTopic() ([]string, error) {
	allRegex := regexp.MustCompile(SubjectAllChannels)
	allTopics, err := ktopics.ListRe(context.Background(), ps.client, allRegex)
	if err != nil {
		return []string{}, err
	}
	var topics []string
	for _, t := range allTopics {
		topics = append(topics, t.Name)
	}
	return topics, nil
}

func (ps *pubsub) checkTopic(topic, id string, err error) (map[string]subscription, error) {
	// Check topic
	s, ok := ps.subscriptions[topic]
	switch ok {
	case true:
		// Check topic ID
		if _, ok := s[id]; ok {
			return map[string]subscription{}, err
		}
	default:
		s = make(map[string]subscription)
		ps.subscriptions[topic] = s
	}
	return s, nil
}

func (ps *pubsub) configReader(id, topic string, s map[string]subscription, handler messaging.MessageHandler) {
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
}
