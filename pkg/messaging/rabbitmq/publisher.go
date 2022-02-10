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
	queueName        = "mainflux-queue"
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
	// defer conn.Close()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// defer ch.Close()
	ret := &publisher{
		connection: conn,
		channel:    ch,
	}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
		return err
	}
	subject := fmt.Sprintf("%s.%s.%s", Exchange, ChansPrefix, topic)
	fmt.Println(subject)
	if err := pub.channel.ExchangeDeclare(subject, ExchangeKind, true, false, false, false, nil); err != nil {
		return err
	}

	err = pub.channel.Publish(
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

func (pub *publisher) Close() {
	pub.connection.Close()
}
