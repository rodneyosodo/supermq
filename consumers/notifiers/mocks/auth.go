// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/policies"
	"google.golang.org/grpc"
)

var _ policies.AuthServiceClient = (*authServiceMock)(nil)

type authServiceMock struct {
	users map[string]string
}

// NewAuth creates mock of auth service.
func NewAuth(users map[string]string) policies.AuthServiceClient {
	return &authServiceMock{users}
}

func (svc authServiceMock) Identify(ctx context.Context, req *policies.IdentifyReq, opts ...grpc.CallOption) (*policies.IdentifyRes, error) {
	if id, ok := svc.users[req.GetToken()]; ok {
		return &policies.IdentifyRes{Id: id}, nil
	}
	return nil, errors.ErrAuthentication
}

func (svc authServiceMock) Issue(ctx context.Context, req *policies.IssueReq, opts ...grpc.CallOption) (*policies.IssueRes, error) {
	if id, ok := svc.users[req.GetIdentity()]; ok {
		return &policies.IssueRes{Token: id}, nil
	}
	return nil, errors.ErrAuthentication
}

func (svc authServiceMock) Authorize(ctx context.Context, req *policies.AuthorizeReq, _ ...grpc.CallOption) (r *policies.AuthorizeRes, err error) {
	panic("not implemented")
}

func (svc authServiceMock) AddPolicy(ctx context.Context, req *policies.AddPolicyReq, opts ...grpc.CallOption) (*policies.AddPolicyRes, error) {
	panic("not implemented")
}

func (svc authServiceMock) DeletePolicy(ctx context.Context, req *policies.DeletePolicyReq, opts ...grpc.CallOption) (*policies.DeletePolicyRes, error) {
	panic("not implemented")
}
