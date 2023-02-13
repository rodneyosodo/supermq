// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"fmt"

	kf "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mainflux/mainflux/pkg/messaging"
	"google.golang.org/protobuf/proto"
)

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	prod *kf.Producer
}

// NewPublisher returns Kafka message Publisher.
func NewPublisher(url string) (messaging.Publisher, error) {
	prod, err := kf.NewProducer(&kf.ConfigMap{"bootstrap.servers": url})
	if err != nil {
		return &publisher{}, err
	}
	ret := &publisher{
		prod: prod,
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

	kafkaMsg := kf.Message{
		TopicPartition: kf.TopicPartition{Topic: &subject},
		Value:          data,
	}

	if err := pub.prod.Produce(&kafkaMsg, nil); err != nil {
		return err
	}

	pub.prod.Flush(1 * 1000)
	return nil
}

func (pub *publisher) Close() error {
	pub.prod.Close()
	return nil
}
