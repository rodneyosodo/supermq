// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/mainflux/mainflux/users/policies"
)

func authorizeEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authReq)

		if err := req.validate(); err != nil {
			return authorizeRes{}, err
		}
		aReq := policies.AccessRequest{Subject: req.subject, Object: req.object, Action: req.action, Entity: req.entityType}
		err := svc.Authorize(ctx, aReq)
		if err != nil {
			return authorizeRes{}, err
		}
		return authorizeRes{authorized: true}, err
	}
}

func issueEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(issueReq)
		if err := req.validate(); err != nil {
			return issueRes{}, err
		}

		tkn, err := svc.IssueToken(ctx, req.identity, req.secret)
		if err != nil {
			return issueRes{}, err
		}

		return issueRes{token: tkn.AccessToken}, nil
	}
}

func identifyEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(identifyReq)
		if err := req.validate(); err != nil {
			return identifyRes{}, err
		}

		id, err := svc.Identify(ctx, req.token)
		if err != nil {
			return identifyRes{}, err
		}

		ret := identifyRes{
			id: id,
		}
		return ret, nil
	}
}

func addPolicyEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addPolicyReq)
		if err := req.validate(); err != nil {
			return addPolicyRes{}, err
		}
		policy := policies.Policy{Subject: req.subject, Object: req.object, Actions: req.action}
		err := svc.AddPolicy(ctx, req.token, policy)
		if err != nil {
			return addPolicyRes{}, err
		}
		return addPolicyRes{added: true}, err
	}
}

func deletePolicyEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(policyReq)
		if err := req.validate(); err != nil {
			return deletePolicyRes{}, err
		}

		policy := policies.Policy{Subject: req.subject, Object: req.object, Actions: []string{req.action}}
		err := svc.DeletePolicy(ctx, req.token, policy)
		if err != nil {
			return deletePolicyRes{}, err
		}
		return deletePolicyRes{deleted: true}, nil
	}
}
