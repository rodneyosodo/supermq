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
		writer: &kafka.Writer{
			Addr:  kafka.TCP(url),
			Async: true,
		},
	}
	return ret, nil

}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	kafkaMsg := kafka.Message{
		Topic: subject,
		Value: data,
	}
	err = pub.writer.WriteMessages(context.TODO(), kafkaMsg)
	if strings.Contains(fmt.Sprint(err), "[5] Leader Not Available:") {
		time.Sleep(1 * time.Second)
		return pub.writer.WriteMessages(context.TODO(), kafkaMsg)
	}
	return err
}

func (pub *publisher) Close() {
	pub.conn.Close()
}
