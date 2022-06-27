// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/segmentio/kafka-go"
)

var (
	_ messaging.Publisher = (*publisher)(nil)
)

type publisher struct {
	url string
}

// NewPublisher returns Kafka message Publisher.
func NewPublisher(url string) (messaging.Publisher, error) {
	ret := &publisher{
		url: url,
	}
	return ret, nil

}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	writer := &kafka.Writer{
		Addr:  kafka.TCP(pub.url),
		Async: true,
		Topic: subject,
	}
	defer writer.Close()

	kafkaMsg := kafka.Message{
		Value: data,
	}
	return writer.WriteMessages(context.Background(), kafkaMsg)
}

func (pub *publisher) Close() error {
	return nil
}
