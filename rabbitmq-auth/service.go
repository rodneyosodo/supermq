// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package rabbitmqauth

import (
	"context"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/policies"
)

type Service interface {
	AuthenticateUser(ctx context.Context, username, password, vhost string) bool
	AuthenticateResource(ctx context.Context, username, vhost string) bool
	AuthorizePubSub(ctx context.Context, username, topic, permission string) bool
}

type service struct {
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
	vhost    string
}

func NewService(clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, vhost string) Service {
	return &service{
		clients:  clients,
		channels: channels,
		vhost:    vhost,
	}
}

func (s *service) AuthenticateUser(ctx context.Context, username, password, vhost string) bool {
	if s.vhost != vhost {
		return false
	}

	resp, err := s.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{ClientSecret: password})
	if err != nil {
		return false
	}
	if !resp.GetAuthenticated() {
		return false
	}
	if resp.GetId() != username {
		return false
	}

	return true
}

func (s *service) AuthenticateResource(ctx context.Context, username, vhost string) bool {
	if s.vhost != vhost {
		return false
	}

	if username == "" {
		return false
	}

	return true
}

func (s *service) AuthorizePubSub(ctx context.Context, username, topic, permission string) bool {
	if permission != "read" && permission != "write" {
		return false
	}
	var perm connections.ConnType
	switch permission {
	case "read":
		perm = connections.Subscribe
	case "write":
		perm = connections.Publish
	}

	channel := strings.Split(topic, ".")[1]

	resp, err := s.channels.Authorize(ctx, &grpcChannelsV1.AuthzReq{
		ClientId:   username,
		ClientType: policies.ClientType,
		Type:       uint32(perm),
		ChannelId:  channel,
	})
	if err != nil {
		return false
	}
	if !resp.GetAuthorized() {
		return false
	}

	return true
}
