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
)

var _ messaging.PubSub = (*pubsub)(nil)

// PubSub wraps messaging Publisher exposing
// Close() method for RabbitMQ connection.
type PubSub interface {
	messaging.PubSub
	Close()
}

type pubsub struct {
	publisher     messaging.Publisher
	conn          *amqp.Connection
	logger        log.Logger
	queue         amqp.Queue
	ch            *amqp.Channel
	subscriptions map[string]bool
	done          chan error
	mutex         sync.Mutex
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
	pub, err := NewPublisher(url)
	if err != nil {
		return nil, err
	}
	ret := &pubsub{
		publisher:     pub,
		conn:          conn,
		queue:         queue,
		ch:            ch,
		logger:        logger,
		subscriptions: make(map[string]bool),
		done:          make(chan error),
	}
	return ret, nil
}

func (ps *pubsub) Publish(topic string, msg messaging.Message) error {
	if topic == "" {
		return errEmptyTopic
	}
	if err := ps.publisher.Publish(topic, msg); err != nil {
		return err
	}

	return nil
}

func (ps *pubsub) Subscribe(topic string, handler messaging.MessageHandler) error {
	if topic == "" {
		return errEmptyTopic
	}
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if _, ok := ps.subscriptions[topic]; ok {
		return errAlreadySubscribed
	}

	subject := fmt.Sprintf("%s.%s.%s", exchange, chansPrefix, topic)

	if err := ps.ch.ExchangeDeclare(subject, exchangeKind, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ps.ch.QueueBind(ps.queue.Name, routingKey, subject, false, nil); err != nil {
		return err
	}
	msgs, err := ps.ch.Consume(ps.queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	go ps.handle(msgs, handler)
	ps.subscriptions[topic] = true

	return nil
}

func (ps *pubsub) Unsubscribe(topic string) error {
	if topic == "" {
		return errEmptyTopic
	}
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if _, ok := ps.subscriptions[topic]; !ok {
		return errNotSubscribed
	}
	subject := fmt.Sprintf("%s.%s.%s", exchange, chansPrefix, topic)
	if err := ps.ch.QueueUnbind(ps.queue.Name, routingKey, subject, nil); err != nil {
		return err
	}

	delete(ps.subscriptions, topic)
	return nil
}

func (ps *pubsub) Close() {
	ps.conn.Close()
}

func (ps *pubsub) handle(deliveries <-chan amqp.Delivery, h messaging.MessageHandler) {
	for d := range deliveries {
		var msg messaging.Message
		if err := proto.Unmarshal(d.Body, &msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h(msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
			return
		}
	}
	ps.done <- nil
}
