// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package kafka holds the implementation of the Publisher and PubSub
// interfaces for the Kafka messaging system, the internal messaging
// broker of the Mainflux IoT platform. Due to the practical requirements
// implementation Publisher is created alongside PubSub. The reason for
// this is that Subscriber implementation of Kafka brings the burden of
// additional struct fields which are not used by Publisher. Subscriber
// is not implemented separately because PubSub can be used where Subscriber is needed.
package kafka
