// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

var _ messaging.Publisher = (*publisher)(nil)

var (
	numPartitions     = 1
	replicationFactor = 1
	batchTimeout      = time.Microsecond
)

type publisher struct {
	url    string
	conn   *kafka.Conn
	mu     sync.Mutex
	topics map[string]*kafka.Writer
}

// NewPublisher returns Kafka message Publisher.
func NewPublisher(url string) (messaging.Publisher, error) {
	conn, err := kafka.Dial("tcp", url)
	if err != nil {
		return &publisher{}, err
	}
	ret := &publisher{
		url:    url,
		conn:   conn,
		topics: make(map[string]*kafka.Writer),
	}
	return ret, nil

}

func (pub *publisher) Publish(topic string, msg *messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}

	kafkaMsg := kafka.Message{
		Value: data,
	}

	writer, ok := pub.topics[subject]
	if ok {
		if err := writer.WriteMessages(context.Background(), kafkaMsg); err != nil {
			return err
		}
		return nil
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             subject,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}
	if err := pub.conn.CreateTopics(topicConfigs...); err != nil {
		return err
	}
	writer = &kafka.Writer{
		Addr:                   kafka.TCP(pub.url),
		Topic:                  subject,
		RequiredAcks:           kafka.RequireAll,
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           batchTimeout,
		AllowAutoTopicCreation: true,
	}
	if err := writer.WriteMessages(context.Background(), kafkaMsg); err != nil {
		return err
	}
	pub.mu.Lock()
	defer pub.mu.Unlock()
	pub.topics[subject] = writer
	return nil
}

func (pub *publisher) Close() error {
	defer pub.conn.Close()

	pub.mu.Lock()
	defer pub.mu.Unlock()

	topics := make([]string, 0, len(pub.topics))
	for topic := range pub.topics {
		topics = append(topics, topic)
		pub.topics[topic].Close()
	}

	req := &kafka.DeleteTopicsRequest{
		Addr:   kafka.TCP(pub.url),
		Topics: topics,
	}
	client := kafka.Client{
		Addr: kafka.TCP(pub.url),
	}
	if _, err := client.DeleteTopics(context.Background(), req); err != nil {
		return err
	}
	return nil
}
