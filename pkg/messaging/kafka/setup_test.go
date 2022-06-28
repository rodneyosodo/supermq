// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package kafka_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/kafka"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var pubsub messaging.PubSub

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "johnnypark/kafka-zookeeper",
		Tag:        "latest",
		Env: []string{
			"ADVERTISED_HOST=127.0.0.1",
			"ADVERTISED_PORT=9092",
		},
		ExposedPorts: []string{
			"9092",
			"2181",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}
	handleInterrupt(pool, container)
	address := fmt.Sprintf("%s:%s", "localhost", container.GetPort("9092/tcp"))

	// As kafka doesn't support a readiness endpoint we have to ensure that kafka is ready before we test it.
	// When you immediately start testing it will throw an EOF error thus we should wait for sometime
	// before starting the tests after bringing the docker container up
	time.Sleep(10 * time.Second)

	logger, err := logger.New(os.Stdout, "error")
	if err != nil {
		log.Fatalf(err.Error())
	}
	if err := pool.Retry(func() error {
		pubsub, err = kafka.NewPubSub(address, "", logger)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()
	if err := pool.Purge(container); err != nil {
		log.Fatalf("Could not purge container: %s", err)
	}

	os.Exit(code)
}

func handleInterrupt(pool *dockertest.Pool, container *dockertest.Resource) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := pool.Purge(container); err != nil {
			log.Fatalf("Could not purge container: %s", err)
		}
		os.Exit(0)
	}()
}
