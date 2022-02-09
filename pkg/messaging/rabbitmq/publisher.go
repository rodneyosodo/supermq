// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

var _ messaging.Publisher = (*publisher)(nil)

var (
	queueName        = "mainflux-mqtt"
	queueDurability  = false
	mandatory        = false
)

type publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// Publisher wraps messaging Publisher exposing
// Close() method for RabbitMQ connection.
type Publisher interface {
	messaging.Publisher
	Close()
}

// NewPublisher returns RabbitMQ message Publisher.
func NewPublisher(url string) (Publisher, error) {
	endpoint := fmt.Sprintf("amqp://%s", url)
	conn, err := amqp.Dial(endpoint)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	ret := &publisher{
		connection: conn,
		channel:    ch,
	}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
		return err
	}

	subject := fmt.Sprintf("%s.%s.%s", exchange, chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s.%s", exchange, subject, msg.Subtopic)
	}
	queue, err := pub.channel.QueueDeclare(
		queueName,
		queueDurability,
		queueDelete,
		queueExclusivity,
		queueWait,
		nil,
	)
	if err != nil {
		return err
	}
	err = pub.channel.Publish(
		subject,
		queue.Name,
		mandatory,
		immediate,
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

func (pub *publisher) Close() {
	pub.connection.Close()
}
