// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/segmentio/kafka-go"
)

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	conn   *kafka.Conn
	url    string
	writer *kafka.Writer
}

// Publisher wraps messaging Publisher exposing
// Close() method for Kafka connection.
type Publisher interface {
	messaging.Publisher
	Close()
}

// NewPublisher returns Kafka message Publisher.
func NewPublisher(url string) (Publisher, error) {
	conn, err := kafka.Dial("tcp", url)
	if err != nil {
		return nil, err
	}
	ret := &publisher{
		conn: conn,
		url:  url,
		writer: &kafka.Writer{
			Addr:  kafka.TCP(url),
			Async: true,
		},
	}
	return ret, nil

}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	if topic == "" {
		return errEmptyTopic
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{pub.url},
		Topic:   subject,
		Async:   true,
	})
	defer writer.Close()

	kafkaMsg := kafka.Message{
		Value: data,
	}
	if err := pub.writer.WriteMessages(context.TODO(), kafkaMsg); err != nil {
		// Sometime it take time for leader to be elected. If that is so, we retry to publish message
		if strings.Contains(fmt.Sprint(err), "[5] Leader Not Available:") {
			time.Sleep(2 * time.Second)
			return pub.writer.WriteMessages(context.TODO(), kafkaMsg)
		}
	}
	return nil
}

func (pub *publisher) Close() {
	pub.conn.Close()
}
