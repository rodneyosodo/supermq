// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/policies"
	"google.golang.org/grpc"
)

var _ policies.AuthServiceClient = (*authServiceClient)(nil)

type authServiceClient struct {
	users map[string]string
}

// NewAuthServiceClient creates mock of auth service.
func NewAuthServiceClient(users map[string]string) policies.AuthServiceClient {
	return &authServiceClient{users}
}

func (svc authServiceClient) Identify(ctx context.Context, req *policies.IdentifyReq, opts ...grpc.CallOption) (*policies.IdentifyRes, error) {
	if id, ok := svc.users[req.GetToken()]; ok {
		return &policies.IdentifyRes{Id: id}, nil
	}
	return nil, errors.ErrAuthentication
}

func (svc *authServiceClient) Issue(ctx context.Context, req *policies.IssueReq, opts ...grpc.CallOption) (*policies.IssueRes, error) {
	if id, ok := svc.users[req.GetIdentity()]; ok {
		return &policies.IssueRes{Token: id}, nil
	}
	return nil, errors.ErrAuthentication
}

func (svc *authServiceClient) Authorize(ctx context.Context, req *policies.AuthorizeReq, _ ...grpc.CallOption) (r *policies.AuthorizeRes, err error) {
	panic("not implemented")
}

func (svc authServiceClient) AddPolicy(ctx context.Context, req *policies.AddPolicyReq, opts ...grpc.CallOption) (*policies.AddPolicyRes, error) {
	panic("not implemented")
}

func (svc authServiceClient) DeletePolicy(ctx context.Context, req *policies.DeletePolicyReq, opts ...grpc.CallOption) (*policies.DeletePolicyRes, error) {
	panic("not implemented")
}
