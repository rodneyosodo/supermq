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
	chansPrefix        = "channels"
	SubjectAllChannels = "channels.>"
	routingKey         = "application specific routing key for fancy topologies"
	exchange           = "mainflux"
	exchangeKind       = "fanout"
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
}

// MainfluxQueue captures the queue values
type MainfluxQueue struct {
	Name        string
	Durability  bool
	Delete      bool
	Exclusivity bool
	Wait        bool
}

// NewPubSub returns RabbitMQ message publisher/subscriber.
func NewPubSub(url string, q MainfluxQueue, logger log.Logger) (PubSub, error) {
	endpoint := fmt.Sprintf("amqp://%s", url)
	conn, err := amqp.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(q.Name, q.Durability, q.Delete, q.Exclusivity, q.Wait, nil)
	if err != nil {
		return nil, err
	}
	ret := &pubsub{
		conn:          conn,
		queue:         queue,
		channel:       ch,
		logger:        logger,
		subscriptions: make(map[string]bool),
	}
	return ret, nil
}

func (ps *pubsub) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s.%s", exchange, chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s.%s", exchange, subject, msg.Subtopic)
	}
	err = ps.channel.Publish(
		subject,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:     amqp.Table{},
			ContentType: "text/plain",
			Priority:    2,
			UserId:      "mainflux_amqp",
			AppId:       "mainflux",
			Body:        []byte(data),
		})

	if err != nil {
		return err
	}

	return nil
}

// nh := ps.natsHandler(handler)

// if ps.queue != "" {
// 	sub, err := ps.conn.QueueSubscribe(topic, ps.queue, nh)

func (ps *pubsub) Subscribe(topic string, handler messaging.MessageHandler) error {
	if topic == "" {
		return errEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if _, ok := ps.subscriptions[topic]; ok {
		return errAlreadySubscribed
	}

	// subject := fmt.Sprintf("%s.%s.%s", exchange, chansPrefix, topic)

	if err := ps.channel.ExchangeDeclare(exchange, exchangeKind, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ps.channel.QueueBind(ps.queue.Name, routingKey, exchange, false, nil); err != nil {
		return err
	}
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

	if err := ps.channel.QueueBind(ps.queue.Name, routingKey, exchange, false, nil); err != nil {
		return err
	}

	delete(ps.subscriptions, topic)
	return nil
}

func (ps *pubsub) Close() {
	ps.conn.Close()
}
