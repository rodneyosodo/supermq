// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/pkg/errors"
	rabbitmqauth "github.com/absmach/supermq/rabbitmq-auth"
	"github.com/go-kit/kit/endpoint"
)

func authenticateEndpoint(svc rabbitmqauth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(authRequest)
		if !ok {
			return authResponse{authenticated: false}, errors.New("invalid request type")
		}

		if err := req.Validate(); err != nil {
			return authResponse{authenticated: false}, errors.Wrap(apiutil.ErrValidation, err)
		}

		if ok := svc.Authenticate(ctx, req.Username, req.Password, req.Vhost); !ok {
			return authResponse{authenticated: false}, errors.ErrAuthentication
		}

		return authResponse{authenticated: true}, nil
	}
}
