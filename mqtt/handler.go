// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/absmach/mgate/pkg/session"
	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
)

var _ session.Handler = (*handler)(nil)

const protocol = "mqtt"

// Log message formats.
const (
	LogInfoSubscribed   = "subscribed with client_id %s to topics %s"
	LogInfoUnsubscribed = "unsubscribed client_id %s from topics %s"
	LogInfoConnected    = "connected with client_id %s"
	LogInfoDisconnected = "disconnected client_id %s and username %s"
	LogInfoPublished    = "published with client_id %s to the topic %s"
)

// Error wrappers for MQTT errors.
var (
	ErrMalformedSubtopic            = errors.New("malformed subtopic")
	ErrClientNotInitialized         = errors.New("client is not initialized")
	ErrMalformedTopic               = errors.New("malformed topic")
	ErrMissingClientID              = errors.New("client_id not found")
	ErrMissingTopicPub              = errors.New("failed to publish due to missing topic")
	ErrMissingTopicSub              = errors.New("failed to subscribe due to missing topic")
	ErrFailedConnect                = errors.New("failed to connect")
	ErrFailedSubscribe              = errors.New("failed to subscribe")
	ErrFailedUnsubscribe            = errors.New("failed to unsubscribe")
	ErrFailedPublish                = errors.New("failed to publish")
	ErrFailedDisconnect             = errors.New("failed to disconnect")
	ErrFailedPublishDisconnectEvent = errors.New("failed to publish disconnect event")
	ErrFailedParseSubtopic          = errors.New("failed to parse subtopic")
	ErrFailedPublishConnectEvent    = errors.New("failed to publish connect event")
	ErrFailedSubscribeEvent         = errors.New("failed to publish subscribe event")
	ErrFailedPublishToMsgBroker     = errors.New("failed to publish to supermq message broker")
)

var (
	errInvalidUserId = errors.New("invalid user id")
	channelRegExp    = regexp.MustCompile(`^\/?channels\/([\w\-]+)\/messages(\/[^?]*)?(\?.*)?$`)
)

// Event implements events.Event interface.
type handler struct {
	pubsub    messaging.PubSub
	publisher messaging.Publisher
	clients   grpcClientsV1.ClientsServiceClient
	channels  grpcChannelsV1.ChannelsServiceClient
	logger    *slog.Logger
}

// NewHandler creates new Handler entity.
func NewHandler(pubsub messaging.PubSub, publisher messaging.Publisher, logger *slog.Logger, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient) session.Handler {
	return &handler{
		logger:    logger,
		pubsub:    pubsub,
		publisher: publisher,
		clients:   clients,
		channels:  channels,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the MQTT broker.
func (h *handler) AuthConnect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}

	if s.ID == "" {
		return ErrMissingClientID
	}

	pwd := string(s.Password)

	res, err := h.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{ClientSecret: pwd})
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if !res.GetAuthenticated() {
		return svcerr.ErrAuthentication
	}

	if s.Username != "" && res.GetId() != s.Username {
		return errInvalidUserId
	}

	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	if topic == nil {
		return ErrMissingTopicPub
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}

	return h.authAccess(ctx, string(s.Username), *topic, connections.Publish)
}

// AuthSubscribe is called on device subscribe,
// prior forwarding to the MQTT broker.
func (h *handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}
	if topics == nil || *topics == nil {
		return ErrMissingTopicSub
	}

	for _, topic := range *topics {
		if err := h.authAccess(ctx, string(s.Username), topic, connections.Subscribe); err != nil {
			return err
		}
	}

	return nil
}

// Connect - after client successfully connected.
func (h *handler) Connect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedConnect, ErrClientNotInitialized)
	}
	h.logger.Info(fmt.Sprintf(LogInfoConnected, s.ID))
	return nil
}

// Publish - after client successfully published.
func (h *handler) Publish(ctx context.Context, topic *string, payload *[]byte) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedPublish, ErrClientNotInitialized)
	}
	h.logger.Info(fmt.Sprintf(LogInfoPublished, s.ID, *topic))
	// Topics are in the format:
	// channels/<channel_id>/messages/<subtopic>/.../ct/<content_type>

	channelParts := channelRegExp.FindStringSubmatch(*topic)
	if len(channelParts) < 2 {
		return errors.Wrap(ErrFailedPublish, ErrMalformedTopic)
	}

	chanID := channelParts[1]
	subtopic := channelParts[2]

	subtopic, err := parseSubtopic(subtopic)
	if err != nil {
		return errors.Wrap(ErrFailedParseSubtopic, err)
	}

	msg := messaging.Message{
		Protocol:  protocol,
		Channel:   chanID,
		Subtopic:  subtopic,
		Publisher: s.Username,
		Payload:   *payload,
		Created:   time.Now().UnixNano(),
	}

	if err := h.pubsub.Publish(ctx, msg.GetChannel(), &msg); err != nil {
		return errors.Wrap(ErrFailedPublishToMsgBroker, err)
	}

	return nil
}

// Subscribe - after client successfully subscribed.
func (h *handler) Subscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedSubscribe, ErrClientNotInitialized)
	}

	if topics == nil || *topics == nil {
		return ErrMissingTopicSub
	}

	for _, topic := range *topics {
		topic := strings.ReplaceAll(topic, "/", ".")
		topic = strings.ReplaceAll(topic, ".messages", "")
		subCfg := messaging.SubscriberConfig{
			ID:    s.Username,
			Topic: topic,
			Handler: forwarder{
				topic:     topic,
				publisher: h.publisher,
			},
		}
		fmt.Printf("Subscriber config %+v\n", subCfg)
		if err := h.pubsub.Subscribe(ctx, subCfg); err != nil {
			return errors.Wrap(ErrFailedSubscribe, err)
		}
	}

	h.logger.Info(fmt.Sprintf(LogInfoSubscribed, s.ID, strings.Join(*topics, ",")))

	return nil
}

// Unsubscribe - after client unsubscribed.
func (h *handler) Unsubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedUnsubscribe, ErrClientNotInitialized)

	}

	for _, topic := range *topics {
		if err := h.pubsub.Unsubscribe(ctx, s.Username, topic); err != nil {
			return errors.Wrap(ErrFailedUnsubscribe, err)
		}
	}

	h.logger.Info(fmt.Sprintf(LogInfoUnsubscribed, s.ID, strings.Join(*topics, ",")))

	return nil
}

// Disconnect - connection with broker or client lost.
func (h *handler) Disconnect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedDisconnect, ErrClientNotInitialized)
	}
	h.logger.Error(fmt.Sprintf(LogInfoDisconnected, s.ID, s.Password))

	return nil
}

func (h *handler) authAccess(ctx context.Context, clientID, topic string, msgType connections.ConnType) error {
	// Topics are in the format:
	// channels/<channel_id>/messages/<subtopic>/.../ct/<content_type>
	if !channelRegExp.MatchString(topic) {
		return ErrMalformedTopic
	}

	channelParts := channelRegExp.FindStringSubmatch(topic)
	if len(channelParts) < 1 {
		return ErrMalformedTopic
	}

	chanID := channelParts[1]

	ar := &grpcChannelsV1.AuthzReq{
		Type:       uint32(msgType),
		ClientId:   clientID,
		ClientType: policies.ClientType,
		ChannelId:  chanID,
	}
	res, err := h.channels.Authorize(ctx, ar)
	if err != nil {
		return err
	}
	if !res.GetAuthorized() {
		return svcerr.ErrAuthorization
	}

	return nil
}

func parseSubtopic(subtopic string) (string, error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", ErrMalformedSubtopic
	}
	subtopic = strings.ReplaceAll(subtopic, "/", ".")

	elems := strings.Split(subtopic, ".")
	filteredElems := []string{}
	for _, elem := range elems {
		if elem == "" {
			continue
		}

		if len(elem) > 1 && (strings.Contains(elem, "*") || strings.Contains(elem, ">")) {
			return "", ErrMalformedSubtopic
		}

		filteredElems = append(filteredElems, elem)
	}

	subtopic = strings.Join(filteredElems, ".")
	return subtopic, nil
}

type forwarder struct {
	topic     string
	logger    *slog.Logger
	publisher messaging.Publisher
}

func (f forwarder) Handle(msg *messaging.Message) error {
	fmt.Printf("supposed to handle message %+v\n", msg)

	if msg.GetProtocol() == protocol {
		return nil
	}

	topic := "channels/" + msg.GetChannel() + "/messages"
	if msg.GetSubtopic() != "" {
		topic = topic + "/" + strings.ReplaceAll(msg.GetSubtopic(), ".", "/")
	}

	go func() {
		if err := f.publisher.Publish(context.Background(), topic, msg); err != nil {
			f.logger.Warn(fmt.Sprintf("Failed to forward message: %s", err))
		}
	}()

	return nil
}

func (p forwarder) Cancel() error {
	return p.publisher.Close()
}
