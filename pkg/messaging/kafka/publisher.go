// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/segmentio/kafka-go"
)

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	url    string
	mu     sync.Mutex
	topics []string
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

	conn, err := kafka.Dial("tcp", pub.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             subject,
			NumPartitions:     -1,
			ReplicationFactor: -1,
		},
	}
	if err := conn.CreateTopics(topicConfigs...); err != nil {
		return err
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(pub.url),
		Async:        true,
		MaxAttempts:  100,
		RequiredAcks: kafka.RequireAll,
	}
	defer writer.Close()

	kafkaMsg := kafka.Message{
		Value: data,
		Topic: subject,
	}
	if err := writer.WriteMessages(context.Background(), kafkaMsg); err != nil {
		return err
	}
	pub.mu.Lock()
	defer pub.mu.Unlock()
	if !pub.topicInTopics(subject) {
		pub.topics = append(pub.topics, subject)
	}
	return nil
}

func (pub *publisher) Close() error {
	pub.mu.Lock()
	defer pub.mu.Unlock()
	req := &kafka.DeleteTopicsRequest{
		Addr:   kafka.TCP(pub.url),
		Topics: pub.topics,
	}
	client := kafka.Client{
		Addr: kafka.TCP(pub.url),
	}
	if _, err := client.DeleteTopics(context.Background(), req); err != nil {
		return err
	}
	return nil
}

func (pub *publisher) topicInTopics(topic string) bool {
	for _, t := range pub.topics {
		if t == topic {
			return true
		}
	}
	return false
}
