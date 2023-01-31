package grpc

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/auth/policies"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ mainflux.AuthServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	issue        kitgrpc.Handler
	identify     kitgrpc.Handler
	authorize    kitgrpc.Handler
	addPolicy    kitgrpc.Handler
	deletePolicy kitgrpc.Handler
	listPolicies kitgrpc.Handler
	assign       kitgrpc.Handler
	members      kitgrpc.Handler
	mainflux.UnimplementedAuthServiceServer
}

// NewServer returns new AuthServiceServer instance.
func NewServer(psvc policies.Service, ksvc keys.Service) mainflux.AuthServiceServer {
	return &grpcServer{
		issue: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("issue"))(issueEndpoint(ksvc)),
			decodeIssueRequest,
			encodeIssueResponse,
		),
		identify: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("identify"))(identifyEndpoint(ksvc)),
			decodeIdentifyRequest,
			encodeIdentifyResponse,
		),
		authorize: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("authorize"))(authorizeEndpoint(psvc)),
			decodeAuthorizeRequest,
			encodeAuthorizeResponse,
		),
		addPolicy: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("add_policy"))(addPolicyEndpoint(psvc)),
			decodeAddPolicyRequest,
			encodeAddPolicyResponse,
		),
		deletePolicy: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("delete_policy"))(deletePolicyEndpoint(psvc)),
			decodeDeletePolicyRequest,
			encodeDeletePolicyResponse,
		),
	}
}

func (s *grpcServer) Issue(ctx context.Context, req *mainflux.IssueReq) (*mainflux.Token, error) {
	_, res, err := s.issue.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.Token), nil
}

func (s *grpcServer) Identify(ctx context.Context, token *mainflux.Token) (*mainflux.UserIdentity, error) {
	_, res, err := s.identify.ServeGRPC(ctx, token)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.UserIdentity), nil
}

func (s *grpcServer) Authorize(ctx context.Context, req *mainflux.AuthorizeReq) (*mainflux.AuthorizeRes, error) {
	_, res, err := s.authorize.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.AuthorizeRes), nil
}

func (s *grpcServer) AddPolicy(ctx context.Context, req *mainflux.AddPolicyReq) (*mainflux.AddPolicyRes, error) {
	_, res, err := s.addPolicy.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.AddPolicyRes), nil
}

func (s *grpcServer) DeletePolicy(ctx context.Context, req *mainflux.DeletePolicyReq) (*mainflux.DeletePolicyRes, error) {
	_, res, err := s.deletePolicy.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.DeletePolicyRes), nil
}

func (s *grpcServer) ListPolicies(ctx context.Context, req *mainflux.ListPoliciesReq) (*mainflux.ListPoliciesRes, error) {
	_, res, err := s.listPolicies.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.ListPoliciesRes), nil
}

func (s *grpcServer) Assign(ctx context.Context, token *mainflux.Assignment) (*empty.Empty, error) {
	_, res, err := s.assign.ServeGRPC(ctx, token)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*empty.Empty), nil
}

func (s *grpcServer) Members(ctx context.Context, req *mainflux.MembersReq) (*mainflux.MembersRes, error) {
	_, res, err := s.members.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.MembersRes), nil
}

func decodeIssueRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.IssueReq)
	return issueReq{id: req.GetId(), email: req.GetEmail(), keyType: req.GetType()}, nil
}

func encodeIssueResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(issueRes)
	return &mainflux.Token{Value: res.value}, nil
}

func decodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.Token)
	return identityReq{token: req.GetValue()}, nil
}

func encodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &mainflux.UserIdentity{Id: res.id, Email: res.email}, nil
}

func decodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.AuthorizeReq)
	return authReq{Act: req.GetAct(), Obj: req.GetObj(), Sub: req.GetSub(), EntityType: req.GetEntityType()}, nil
}

func encodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(authorizeRes)
	return &mainflux.AuthorizeRes{Authorized: res.authorized}, nil
}

func decodeAddPolicyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.AddPolicyReq)
	return policyReq{Sub: req.GetSub(), Obj: req.GetObj(), Act: req.GetAct()}, nil
}

func encodeAddPolicyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(addPolicyRes)
	return &mainflux.AddPolicyRes{Authorized: res.authorized}, nil
}

func decodeDeletePolicyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.DeletePolicyReq)
	return policyReq{Sub: req.GetSub(), Obj: req.GetObj(), Act: req.GetAct()}, nil
}

func encodeDeletePolicyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(deletePolicyRes)
	return &mainflux.DeletePolicyRes{Deleted: res.deleted}, nil
}

func encodeError(err error) error {
	switch {
	case errors.Contains(err, nil):
		return nil
	case errors.Contains(err, errors.ErrMalformedEntity),
		err == apiutil.ErrInvalidAuthKey,
		err == apiutil.ErrMissingID,
		err == apiutil.ErrMissingMemberType,
		err == apiutil.ErrMissingPolicySub,
		err == apiutil.ErrMissingPolicyObj,
		err == apiutil.ErrMissingPolicyAct,
		err == apiutil.ErrMalformedPolicy,
		err == apiutil.ErrMissingPolicyOwner,
		err == apiutil.ErrHigherPolicyRank:
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Contains(err, errors.ErrAuthentication),
		errors.Contains(err, keys.ErrKeyExpired),
		err == apiutil.ErrMissingEmail,
		err == apiutil.ErrBearerToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Contains(err, errors.ErrAuthorization):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
