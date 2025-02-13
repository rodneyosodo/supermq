// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package rabbitmqauth

import (
	"context"

	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
)

type Service interface {
	Authenticate(ctx context.Context, username, password, vhost string) bool
}

type service struct {
	clients grpcClientsV1.ClientsServiceClient
	vhost   string
}

func NewService(clients grpcClientsV1.ClientsServiceClient, vhost string) Service {
	return &service{
		clients: clients,
		vhost:   vhost,
	}
}

func (s *service) Authenticate(ctx context.Context, username, password, vhost string) bool {
	if password != "" && username != "" {
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
	if username != "" && password == "" {
		if s.vhost == vhost {
			return true
		}
	}

	return false
}
