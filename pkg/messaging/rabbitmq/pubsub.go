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

// SubjectAllChannels represents subject to subscribe for all the channels.
const (
	ChansPrefix        = "channels"
	SubjectAllChannels = "channels.>"
	RoutingKey         = "application specific routing key for fancy topologies"
	Exchange           = "mainflux"
	ExchangeKind       = "fanout"
	QueueName          = "mainflux"
	QueueDurability    = true
	QueueDelete        = false
	QueueExclusivity   = true
	QueueWait          = false
	ConsumerTag        = "mainflux-consumer"
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
	conn          *amqp.Connection
	logger        log.Logger
	mu            sync.Mutex
	queue         amqp.Queue
	channel       *amqp.Channel
	subscriptions map[string]bool
	done          chan bool
}

// NewPubSub returns RabbitMQ message publisher/subscriber.
func NewPubSub(url string, logger log.Logger) (PubSub, error) {
	endpoint := fmt.Sprintf("amqp://%s", url)
	conn, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(QueueName, QueueDurability, QueueDelete, QueueExclusivity, QueueWait, nil)
	if err != nil {
		return nil, err
	}
	ret := &pubsub{
		conn:          conn,
		queue:         queue,
		channel:       ch,
		logger:        logger,
		subscriptions: make(map[string]bool),
		done:          make(chan bool),
	}
	return ret, nil
}

func (ps *pubsub) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("%s.%s.%s", Exchange, ChansPrefix, topic)
	if err := ps.channel.ExchangeDeclare(subject, ExchangeKind, true, false, false, false, nil); err != nil {
		return err
	}
	err = ps.channel.Publish(
		subject,
		RoutingKey,
		mandatory,
		immediate,
		amqp.Publishing{
			Headers:     amqp.Table{},
			ContentType: "text/plain",
			Priority:    2,
			AppId:       "mainflux",
			Body:        []byte(data),
		})

	if err != nil {
		return err
	}

	return nil
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

	subject := fmt.Sprintf("%s.%s.%s", Exchange, ChansPrefix, topic)

	if err := ps.channel.ExchangeDeclare(subject, ExchangeKind, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ps.channel.QueueBind(ps.queue.Name, RoutingKey, subject, false, nil); err != nil {
		return err
	}
	msgs, err := ps.channel.Consume(ps.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		ps.handle(msgs, handler)
		wg.Done()
	}()

	ps.subscriptions[topic] = true

	return nil
}

func (ps *pubsub) Unsubscribe(topic string) error {
	if topic == "" {
		return errEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, ok := ps.subscriptions[topic]
	if !ok {
		return errNotSubscribed
	}
	subject := fmt.Sprintf("%s.%s.%s", Exchange, ChansPrefix, topic)
	if err := ps.channel.QueueBind(ps.queue.Name, RoutingKey, subject, false, nil); err != nil {
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
		}
	}
}
