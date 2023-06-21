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

type MockSubjectSet struct {
	Object   string
	Relation []string
}

type authServiceMock struct {
	users    map[string]string
	policies map[string][]MockSubjectSet
}

// NewAuthService creates mock of users service.
func NewAuthService(users map[string]string, policies map[string][]MockSubjectSet) policies.AuthServiceClient {
	return &authServiceMock{users, policies}
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
	for _, policy := range svc.policies[req.GetSubject()] {
		for _, r := range policy.Relation {
			if r == req.GetAction() && policy.Object == req.GetObject() {
				return &policies.AuthorizeRes{Authorized: true}, nil
			}
		}

	}
	return nil, errors.ErrAuthorization
}

func (svc authServiceMock) AddPolicy(ctx context.Context, req *policies.AddPolicyReq, opts ...grpc.CallOption) (*policies.AddPolicyRes, error) {
	if len(req.GetAction()) == 0 || req.GetObject() == "" || req.GetSubject() == "" {
		return &policies.AddPolicyRes{}, errors.ErrMalformedEntity
	}

	obj := req.GetObject()
	svc.policies[req.GetSubject()] = append(svc.policies[req.GetSubject()], MockSubjectSet{Object: obj, Relation: req.GetAction()})
	return &policies.AddPolicyRes{Added: true}, nil
}

func (svc authServiceMock) DeletePolicy(ctx context.Context, in *policies.DeletePolicyReq, opts ...grpc.CallOption) (*policies.DeletePolicyRes, error) {
	// Not implemented yet
	return &policies.DeletePolicyRes{Deleted: true}, nil
}
