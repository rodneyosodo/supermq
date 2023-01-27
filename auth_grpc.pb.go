// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: auth.proto

package mainflux

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ThingsServiceClient is the client API for ThingsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ThingsServiceClient interface {
	CanAccessByKey(ctx context.Context, in *AccessByKeyReq, opts ...grpc.CallOption) (*ThingID, error)
	IsChannelOwner(ctx context.Context, in *ChannelOwnerReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CanAccessByID(ctx context.Context, in *AccessByIDReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Identify(ctx context.Context, in *Token, opts ...grpc.CallOption) (*ThingID, error)
}

type thingsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewThingsServiceClient(cc grpc.ClientConnInterface) ThingsServiceClient {
	return &thingsServiceClient{cc}
}

func (c *thingsServiceClient) CanAccessByKey(ctx context.Context, in *AccessByKeyReq, opts ...grpc.CallOption) (*ThingID, error) {
	out := new(ThingID)
	err := c.cc.Invoke(ctx, "/mainflux.ThingsService/CanAccessByKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) IsChannelOwner(ctx context.Context, in *ChannelOwnerReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/mainflux.ThingsService/IsChannelOwner", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) CanAccessByID(ctx context.Context, in *AccessByIDReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/mainflux.ThingsService/CanAccessByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) Identify(ctx context.Context, in *Token, opts ...grpc.CallOption) (*ThingID, error) {
	out := new(ThingID)
	err := c.cc.Invoke(ctx, "/mainflux.ThingsService/Identify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ThingsServiceServer is the server API for ThingsService service.
// All implementations must embed UnimplementedThingsServiceServer
// for forward compatibility
type ThingsServiceServer interface {
	CanAccessByKey(context.Context, *AccessByKeyReq) (*ThingID, error)
	IsChannelOwner(context.Context, *ChannelOwnerReq) (*emptypb.Empty, error)
	CanAccessByID(context.Context, *AccessByIDReq) (*emptypb.Empty, error)
	Identify(context.Context, *Token) (*ThingID, error)
	mustEmbedUnimplementedThingsServiceServer()
}

// UnimplementedThingsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedThingsServiceServer struct {
}

func (UnimplementedThingsServiceServer) CanAccessByKey(context.Context, *AccessByKeyReq) (*ThingID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CanAccessByKey not implemented")
}
func (UnimplementedThingsServiceServer) IsChannelOwner(context.Context, *ChannelOwnerReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsChannelOwner not implemented")
}
func (UnimplementedThingsServiceServer) CanAccessByID(context.Context, *AccessByIDReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CanAccessByID not implemented")
}
func (UnimplementedThingsServiceServer) Identify(context.Context, *Token) (*ThingID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Identify not implemented")
}
func (UnimplementedThingsServiceServer) mustEmbedUnimplementedThingsServiceServer() {}

// UnsafeThingsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ThingsServiceServer will
// result in compilation errors.
type UnsafeThingsServiceServer interface {
	mustEmbedUnimplementedThingsServiceServer()
}

func RegisterThingsServiceServer(s grpc.ServiceRegistrar, srv ThingsServiceServer) {
	s.RegisterService(&ThingsService_ServiceDesc, srv)
}

func _ThingsService_CanAccessByKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccessByKeyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).CanAccessByKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.ThingsService/CanAccessByKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).CanAccessByKey(ctx, req.(*AccessByKeyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_IsChannelOwner_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChannelOwnerReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).IsChannelOwner(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.ThingsService/IsChannelOwner",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).IsChannelOwner(ctx, req.(*ChannelOwnerReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_CanAccessByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccessByIDReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).CanAccessByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.ThingsService/CanAccessByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).CanAccessByID(ctx, req.(*AccessByIDReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_Identify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Token)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).Identify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.ThingsService/Identify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).Identify(ctx, req.(*Token))
	}
	return interceptor(ctx, in, info, handler)
}

// ThingsService_ServiceDesc is the grpc.ServiceDesc for ThingsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ThingsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mainflux.ThingsService",
	HandlerType: (*ThingsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CanAccessByKey",
			Handler:    _ThingsService_CanAccessByKey_Handler,
		},
		{
			MethodName: "IsChannelOwner",
			Handler:    _ThingsService_IsChannelOwner_Handler,
		},
		{
			MethodName: "CanAccessByID",
			Handler:    _ThingsService_CanAccessByID_Handler,
		},
		{
			MethodName: "Identify",
			Handler:    _ThingsService_Identify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth.proto",
}

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthServiceClient interface {
	Issue(ctx context.Context, in *IssueReq, opts ...grpc.CallOption) (*Token, error)
	Identify(ctx context.Context, in *Token, opts ...grpc.CallOption) (*UserIdentity, error)
	Authorize(ctx context.Context, in *AuthorizeReq, opts ...grpc.CallOption) (*AuthorizeRes, error)
	AddPolicy(ctx context.Context, in *AddPolicyReq, opts ...grpc.CallOption) (*AddPolicyRes, error)
	DeletePolicy(ctx context.Context, in *DeletePolicyReq, opts ...grpc.CallOption) (*DeletePolicyRes, error)
	ListPolicies(ctx context.Context, in *ListPoliciesReq, opts ...grpc.CallOption) (*ListPoliciesRes, error)
	Assign(ctx context.Context, in *Assignment, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Members(ctx context.Context, in *MembersReq, opts ...grpc.CallOption) (*MembersRes, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Issue(ctx context.Context, in *IssueReq, opts ...grpc.CallOption) (*Token, error) {
	out := new(Token)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/Issue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Identify(ctx context.Context, in *Token, opts ...grpc.CallOption) (*UserIdentity, error) {
	out := new(UserIdentity)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/Identify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Authorize(ctx context.Context, in *AuthorizeReq, opts ...grpc.CallOption) (*AuthorizeRes, error) {
	out := new(AuthorizeRes)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/Authorize", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) AddPolicy(ctx context.Context, in *AddPolicyReq, opts ...grpc.CallOption) (*AddPolicyRes, error) {
	out := new(AddPolicyRes)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/AddPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) DeletePolicy(ctx context.Context, in *DeletePolicyReq, opts ...grpc.CallOption) (*DeletePolicyRes, error) {
	out := new(DeletePolicyRes)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/DeletePolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) ListPolicies(ctx context.Context, in *ListPoliciesReq, opts ...grpc.CallOption) (*ListPoliciesRes, error) {
	out := new(ListPoliciesRes)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/ListPolicies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Assign(ctx context.Context, in *Assignment, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/Assign", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Members(ctx context.Context, in *MembersReq, opts ...grpc.CallOption) (*MembersRes, error) {
	out := new(MembersRes)
	err := c.cc.Invoke(ctx, "/mainflux.AuthService/Members", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations must embed UnimplementedAuthServiceServer
// for forward compatibility
type AuthServiceServer interface {
	Issue(context.Context, *IssueReq) (*Token, error)
	Identify(context.Context, *Token) (*UserIdentity, error)
	Authorize(context.Context, *AuthorizeReq) (*AuthorizeRes, error)
	AddPolicy(context.Context, *AddPolicyReq) (*AddPolicyRes, error)
	DeletePolicy(context.Context, *DeletePolicyReq) (*DeletePolicyRes, error)
	ListPolicies(context.Context, *ListPoliciesReq) (*ListPoliciesRes, error)
	Assign(context.Context, *Assignment) (*emptypb.Empty, error)
	Members(context.Context, *MembersReq) (*MembersRes, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct {
}

func (UnimplementedAuthServiceServer) Issue(context.Context, *IssueReq) (*Token, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Issue not implemented")
}
func (UnimplementedAuthServiceServer) Identify(context.Context, *Token) (*UserIdentity, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Identify not implemented")
}
func (UnimplementedAuthServiceServer) Authorize(context.Context, *AuthorizeReq) (*AuthorizeRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authorize not implemented")
}
func (UnimplementedAuthServiceServer) AddPolicy(context.Context, *AddPolicyReq) (*AddPolicyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddPolicy not implemented")
}
func (UnimplementedAuthServiceServer) DeletePolicy(context.Context, *DeletePolicyReq) (*DeletePolicyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePolicy not implemented")
}
func (UnimplementedAuthServiceServer) ListPolicies(context.Context, *ListPoliciesReq) (*ListPoliciesRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPolicies not implemented")
}
func (UnimplementedAuthServiceServer) Assign(context.Context, *Assignment) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Assign not implemented")
}
func (UnimplementedAuthServiceServer) Members(context.Context, *MembersReq) (*MembersRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Members not implemented")
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Issue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssueReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Issue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/Issue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Issue(ctx, req.(*IssueReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Identify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Token)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Identify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/Identify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Identify(ctx, req.(*Token))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Authorize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthorizeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Authorize(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/Authorize",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Authorize(ctx, req.(*AuthorizeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_AddPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddPolicyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).AddPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/AddPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).AddPolicy(ctx, req.(*AddPolicyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_DeletePolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletePolicyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).DeletePolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/DeletePolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).DeletePolicy(ctx, req.(*DeletePolicyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_ListPolicies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPoliciesReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).ListPolicies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/ListPolicies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).ListPolicies(ctx, req.(*ListPoliciesReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Assign_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Assignment)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Assign(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/Assign",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Assign(ctx, req.(*Assignment))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Members_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MembersReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Members(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mainflux.AuthService/Members",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Members(ctx, req.(*MembersReq))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mainflux.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Issue",
			Handler:    _AuthService_Issue_Handler,
		},
		{
			MethodName: "Identify",
			Handler:    _AuthService_Identify_Handler,
		},
		{
			MethodName: "Authorize",
			Handler:    _AuthService_Authorize_Handler,
		},
		{
			MethodName: "AddPolicy",
			Handler:    _AuthService_AddPolicy_Handler,
		},
		{
			MethodName: "DeletePolicy",
			Handler:    _AuthService_DeletePolicy_Handler,
		},
		{
			MethodName: "ListPolicies",
			Handler:    _AuthService_ListPolicies_Handler,
		},
		{
			MethodName: "Assign",
			Handler:    _AuthService_Assign_Handler,
		},
		{
			MethodName: "Members",
			Handler:    _AuthService_Members_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth.proto",
}
