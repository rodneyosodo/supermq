package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	rabbitmq "github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
)

const (
	topic   = "topic"
	channel = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
)

func main() {
	logger, err := logger.New(os.Stdout, "error")
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubsub, err := rabbitmq.NewPubSub("guest:guest@localhost:5672/", logger)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var count int
	for {
		message := fmt.Sprintf("Payload %d", count)
		data := []byte(message)

		expectedMsg := messaging.Message{
			Channel:  channel,
			Subtopic: "demo",
			Payload:  data,
		}
		fmt.Printf("Publish %v\n", count)
		_ = pubsub.Publish(topic, expectedMsg)
		count = count + 1
		time.Sleep(time.Second * 1)
	}

}
