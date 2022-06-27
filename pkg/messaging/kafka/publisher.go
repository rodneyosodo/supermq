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

var (
	_               messaging.Publisher = (*publisher)(nil)
	backoffSchedule                     = []time.Duration{1 * time.Second, 3 * time.Second, 10 * time.Second}
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

	// Ensuring the kafka topic is created after bringing up the cluster.
	for _, backoff := range backoffSchedule {
		err := writer.WriteMessages(context.Background(), kafkaMsg)
		if !strings.Contains(fmt.Sprint(err), "[5] Leader Not Available:") {
			return err
		}
		if err == nil {
			break
		}
		time.Sleep(backoff)
	}
	return nil
}

func (pub *publisher) Close() error {
	return nil
}
