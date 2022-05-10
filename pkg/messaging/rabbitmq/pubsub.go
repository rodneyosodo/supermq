// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	chansPrefix = "channels"
	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.>"
	routingKey         = "mainfluxkey"
	exchange           = "mainflux"
	exchangeKind       = "fanout"
	queueDurability    = true
	queueDelete        = false
	queueExclusivity   = false
	queueWait          = false
	consumerTag        = "mainflux-consumer"
	mandatory          = false
	immediate          = false
)

var (
	errAlreadySubscribed = errors.New("already subscribed to topic")
	errNotSubscribed     = errors.New("not subscribed")
	errEmptyTopic        = errors.New("empty topic")
	errEmptyID           = errors.New("empty ID")
)

var _ messaging.PubSub = (*pubsub)(nil)

// PubSub wraps messaging Publisher exposing
// Close() method for RabbitMQ connection.
type PubSub interface {
	messaging.PubSub
	Close()
}

type subscription struct {
	cancel func() error
}
type pubsub struct {
	publisher
	logger        log.Logger
	queue         amqp.Queue
	subscriptions map[string]map[string]subscription
	mu            sync.Mutex
}

// NewPubSub returns RabbitMQ message publisher/subscriber.
func NewPubSub(url, queueName string, logger log.Logger) (PubSub, error) {
	endpoint := fmt.Sprintf("amqp://%s", url)
	conn, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	queue, err := ch.QueueDeclare(queueName, queueDurability, queueDelete, queueExclusivity, queueWait, nil)
	if err != nil {
		return nil, err
	}
	ret := &pubsub{
		publisher: publisher{
			conn: conn,
			ch:   ch,
		},
		queue:         queue,
		logger:        logger,
		subscriptions: make(map[string]map[string]subscription),
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

	subject := fmt.Sprintf("%s.%s", exchange, topic)

	if err := ps.ch.ExchangeDeclare(subject, exchangeKind, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ps.ch.QueueBind(ps.queue.Name, routingKey, subject, false, nil); err != nil {
		return err
	}
	msgs, err := ps.ch.Consume(ps.queue.Name, id, true, false, false, false, nil)
	if err != nil {
		return err
	}

	go ps.handle(msgs, handler)
	s[id] = subscription{
		cancel: handler.Cancel,
	}

	return nil
}

func (ps *pubsub) Unsubscribe(id, topic string) error {
	defer ps.ch.Cancel(id, false)
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
	current, ok := s[id]
	if !ok {
		return errNotSubscribed
	}
	subject := fmt.Sprintf("%s.%s", exchange, topic)
	if err := ps.ch.QueueUnbind(ps.queue.Name, routingKey, subject, nil); err != nil {
		return err
	}
	if current.cancel != nil {
		if err := current.cancel(); err != nil {
			return err
		}
	}

	delete(s, id)
	if len(s) == 0 {
		delete(ps.subscriptions, topic)
	}
	return nil
}

func (ps *pubsub) handle(deliveries <-chan amqp.Delivery, h messaging.MessageHandler) {
	for d := range deliveries {
		var msg messaging.Message
		if err := proto.Unmarshal(d.Body, &msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h.Handle(msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
			return
		}
	}
}
