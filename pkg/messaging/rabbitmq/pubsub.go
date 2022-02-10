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
	QueueDurability    = false
	QueueDelete        = false
	QueueExclusivity   = true
	QueueWait          = false
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
			// UserId:      "mainflux_amqp",
			AppId: "mainflux",
			Body:  []byte(data),
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

// func (ps *pubsub) ReadMessages(topic string) error {
// 	if topic == "" {
// 		return errEmptyTopic
// 	}
// 	subject := fmt.Sprintf("%s.%s.%s", Exchange, ChansPrefix, topic)

// 	msgs, err := ps.channel.Consume(ps.queue.Name, "", true, false, false, false, nil)
// 	if err != nil {
// 		return err
// 	}

// 	forever := make(chan bool)

// 	go func() {
// 		for d := range msgs {
// 			fmt.Println(fmt.Sprintf(" [x] %s", d.Body))
// 		}
// 	}()

// 	fmt.Println(" [*] Waiting for logs. To exit press CTRL+C")
// 	<-forever
// 	return nil
// }
func (ps *pubsub) Close() {
	ps.conn.Close()
}
