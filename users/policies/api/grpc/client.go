// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/users/policies"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
	"google.golang.org/grpc"
)

const svcName = "mainflux.users.policies.AuthService"

var _ policies.AuthServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	authorize    endpoint.Endpoint
	issue        endpoint.Endpoint
	identify     endpoint.Endpoint
	addPolicy    endpoint.Endpoint
	deletePolicy endpoint.Endpoint
	timeout      time.Duration
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) policies.AuthServiceClient {
	return &grpcClient{
		authorize: otelkit.EndpointMiddleware(otelkit.WithOperation("authorize"))(kitgrpc.NewClient(
			conn,
			svcName,
			"Authorize",
			encodeAuthorizeRequest,
			decodeAuthorizeResponse,
			policies.AuthorizeRes{},
		).Endpoint()),
		issue: otelkit.EndpointMiddleware(otelkit.WithOperation("issue"))(kitgrpc.NewClient(
			conn,
			svcName,
			"Issue",
			encodeIssueRequest,
			decodeIssueResponse,
			policies.IssueRes{},
		).Endpoint()),
		identify: otelkit.EndpointMiddleware(otelkit.WithOperation("identify"))(kitgrpc.NewClient(
			conn,
			svcName,
			"Identify",
			encodeIdentifyRequest,
			decodeIdentifyResponse,
			policies.IdentifyRes{},
		).Endpoint()),
		addPolicy: otelkit.EndpointMiddleware(otelkit.WithOperation("add_policy"))(kitgrpc.NewClient(
			conn,
			svcName,
			"AddPolicy",
			encodeAddPolicyRequest,
			decodeAddPolicyResponse,
			policies.AddPolicyRes{},
		).Endpoint()),
		deletePolicy: otelkit.EndpointMiddleware(otelkit.WithOperation("delete_policy"))(kitgrpc.NewClient(
			conn,
			svcName,
			"DeletePolicy",
			encodeDeletePolicyRequest,
			decodeDeletePolicyResponse,
			policies.DeletePolicyRes{},
		).Endpoint()),

		timeout: timeout,
	}
}

func (client grpcClient) Authorize(ctx context.Context, req *policies.AuthorizeReq, _ ...grpc.CallOption) (r *policies.AuthorizeRes, err error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()
	areq := authReq{subject: req.GetSubject(), object: req.GetObject(), action: req.GetAction(), entityType: req.GetEntityType()}
	res, err := client.authorize(ctx, areq)
	if err != nil {
		return &policies.AuthorizeRes{}, err
	}

	ares := res.(authorizeRes)
	return &policies.AuthorizeRes{Authorized: ares.authorized}, err
}

func decodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.AuthorizeRes)
	return authorizeRes{authorized: res.GetAuthorized()}, nil
}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(authReq)
	return &policies.AuthorizeReq{
		Subject:    req.subject,
		Object:     req.object,
		Action:     req.action,
		EntityType: req.entityType,
	}, nil
}

func (client grpcClient) Issue(ctx context.Context, req *policies.IssueReq, _ ...grpc.CallOption) (*policies.IssueRes, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()
	ireq := issueReq{identity: req.GetIdentity(), secret: req.GetSecret()}
	res, err := client.issue(ctx, ireq)
	if err != nil {
		return nil, err
	}

	ires := res.(issueRes)
	return &policies.IssueRes{Token: ires.token}, nil
}

func encodeIssueRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(issueReq)
	return &policies.IssueReq{Identity: req.identity, Secret: req.secret}, nil
}

func decodeIssueResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.IssueRes)
	return issueRes{token: res.GetToken()}, nil
}

func (client grpcClient) Identify(ctx context.Context, req *policies.IdentifyReq, _ ...grpc.CallOption) (*policies.IdentifyRes, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()

	ireq, err := client.identify(ctx, identifyReq{token: req.GetToken()})
	if err != nil {
		return nil, err
	}

	ires := ireq.(identifyRes)
	return &policies.IdentifyRes{Id: ires.id}, nil
}

func encodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(identifyReq)
	return &policies.IdentifyReq{Token: req.token}, nil
}

func decodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.IdentifyRes)
	return identifyRes{id: res.GetId()}, nil
}

func (client grpcClient) AddPolicy(ctx context.Context, req *policies.AddPolicyReq, opts ...grpc.CallOption) (*policies.AddPolicyRes, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()
	areq := addPolicyReq{token: req.GetToken(), subject: req.GetSubject(), object: req.GetObject(), action: req.GetAction()}
	res, err := client.addPolicy(ctx, areq)
	if err != nil {
		return &policies.AddPolicyRes{}, err
	}

	ares := res.(addPolicyRes)
	return &policies.AddPolicyRes{Added: ares.added}, err
}

func decodeAddPolicyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.AddPolicyRes)
	return addPolicyRes{added: res.GetAdded()}, nil
}

func encodeAddPolicyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(addPolicyReq)
	return &policies.AddPolicyReq{
		Token:   req.token,
		Subject: req.subject,
		Object:  req.object,
		Action:  req.action,
	}, nil
}

func (client grpcClient) DeletePolicy(ctx context.Context, req *policies.DeletePolicyReq, opts ...grpc.CallOption) (*policies.DeletePolicyRes, error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()
	preq := policyReq{token: req.GetToken(), subject: req.GetSubject(), object: req.GetObject()}
	res, err := client.deletePolicy(ctx, preq)
	if err != nil {
		return &policies.DeletePolicyRes{}, err
	}

	pres := res.(deletePolicyRes)
	return &policies.DeletePolicyRes{Deleted: pres.deleted}, err
}

func decodeDeletePolicyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.DeletePolicyRes)
	return deletePolicyRes{deleted: res.GetDeleted()}, nil
}

func encodeDeletePolicyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(policyReq)
	return &policies.DeletePolicyReq{
		Token:   req.token,
		Subject: req.subject,
		Object:  req.object,
	}, nil
}
