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
	topic = "topic"
)

var (
	msgChan = make(chan messaging.Message, 1000)
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
	_ = pubsub.Subscribe(topic, handler)
	time.Sleep(time.Second * 10)
	_ = pubsub.Unsubscribe(topic)
	time.Sleep(time.Second * 10)
	_ = pubsub.Subscribe(topic, handler)
	time.Sleep(time.Second * 10)
	_ = pubsub.Unsubscribe(topic)
	time.Sleep(time.Second * 10)

	// for _ = range msgChan {
	// 	msg := <-msgChan
	// 	fmt.Println(msg)
	// }

}

func handler(msg messaging.Message) error {
	fmt.Printf("Message published: %s\n", string(msg.Payload))
	msgChan <- msg
	return nil
}
