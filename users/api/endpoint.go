// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/users"
)

func registrationEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createUserReq)
		if err := req.validate(); err != nil {
			return createUserRes{}, err
		}
		user, err := svc.Register(ctx, req.token, req.user)
		if err != nil {
			return createUserRes{}, err
		}
		ucr := createUserRes{
			User:    user,
			created: true,
		}

		return ucr, nil
	}
}

func viewUserEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewUserReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		user, err := svc.ViewUser(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return viewUserRes{User: user}, nil
	}
}

func viewProfileEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewUserReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		user, err := svc.ViewProfile(ctx, req.token)
		if err != nil {
			return nil, err
		}
		return viewUserRes{User: user}, nil
	}
}

// Password reset request endpoint.
// When successful password reset link is generated.
// Link is generated using MF_TOKEN_RESET_ENDPOINT env.
// and value from Referer header for host.
// {Referer}+{MF_TOKEN_RESET_ENDPOINT}+{token=TOKEN}
// http://mainflux.com/reset-request?token=xxxxxxxxxxx.
// Email with a link is being sent to the user.
// When user clicks on a link it should get the ui with form to
// enter new password, when form is submitted token and new password
// must be sent as PUT request to 'password/reset' passwordResetEndpoint
func passwordResetRequestEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwResetReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := passwResetReqRes{}
		email := req.Email
		if err := svc.GenerateResetToken(ctx, email, req.Host); err != nil {
			return nil, err
		}
		res.Msg = MailSent

		return res, nil
	}
}

// This is endpoint that actually sets new password in password reset flow.
// When user clicks on a link in email finally ends on this endpoint as explained in
// the comment above.
func passwordResetEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetTokenReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := passwChangeRes{}
		if err := svc.ResetPassword(ctx, req.Token, req.Password); err != nil {
			return nil, err
		}
		return res, nil
	}
}

func listClientsEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listUsersReq)
		if err := req.validate(); err != nil {
			return users.UsersPage{}, err
		}

		pm := users.Page{
			SharedBy: req.sharedBy,
			Status:   req.status,
			Offset:   req.offset,
			Limit:    req.limit,
			OwnerID:  req.owner,
			Name:     req.name,
			Tag:      req.tag,
			Metadata: req.metadata,
		}
		page, err := svc.ListUsers(ctx, req.token, pm)
		if err != nil {
			return users.UsersPage{}, err
		}

		res := userssPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Users: []viewUserRes{},
		}
		for _, user := range page.Users {
			res.Users = append(res.Users, viewUserRes{User: user})
		}

		return res, nil
	}
}

func listMembersEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMembersReq)
		if err := req.validate(); err != nil {
			return memberPageRes{}, err
		}
		page, err := svc.ListMembers(ctx, req.token, req.groupID, req.Page)
		if err != nil {
			return memberPageRes{}, err
		}
		return buildMembersResponse(page), nil
	}
}

func updateUserEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		user := users.User{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		user, err := svc.UpdateUser(ctx, req.token, user)
		if err != nil {
			return nil, err
		}
		return updateUserRes{User: user}, nil
	}
}

func updateUserTagsEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserTagsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		user := users.User{
			ID:   req.id,
			Tags: req.Tags,
		}
		user, err := svc.UpdateUserTags(ctx, req.token, user)
		if err != nil {
			return nil, err
		}
		return updateUserRes{User: user}, nil
	}
}

func updateUserIdentityEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserCredentialsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		user, err := svc.UpdateUserIdentity(ctx, req.token, req.id, req.Identity)
		if err != nil {
			return nil, err
		}
		return updateUserRes{User: user}, nil
	}
}

func changePasswordEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserCredentialsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		user, err := svc.ChangePassword(ctx, req.token, req.OldSecret, req.NewSecret)
		if err != nil {
			return nil, err
		}
		return updateUserRes{User: user}, nil
	}
}

func updateUserOwnerEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateUserOwnerReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		user := users.User{
			ID:    req.id,
			Owner: req.Owner,
		}

		user, err := svc.UpdateUserOwner(ctx, req.token, user)
		if err != nil {
			return nil, err
		}
		return updateUserRes{User: user}, nil
	}
}

func loginEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(loginUserReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		user := users.User{
			Credentials: users.Credentials{
				Identity: req.Identity,
				Secret:   req.Secret,
			},
		}
		token, err := svc.Login(ctx, user)
		if err != nil {
			return nil, err
		}

		return tokenRes{
			AccessToken: token,
		}, nil
	}
}

func enableUserEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeUserStatusReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		user, err := svc.EnableUser(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return deleteUserRes{User: user}, nil
	}
}

func disableUserEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeUserStatusReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		user, err := svc.DisableUser(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return deleteUserRes{User: user}, nil
	}
}

func buildMembersResponse(cp users.MembersPage) memberPageRes {
	res := memberPageRes{
		pageRes: pageRes{
			Total:  cp.Total,
			Offset: cp.Offset,
			Limit:  cp.Limit,
		},
		Members: []viewMembersRes{},
	}
	for _, user := range cp.Members {
		res.Members = append(res.Members, viewMembersRes{User: user})
	}
	return res
}
