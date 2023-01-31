package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// MakeClientsHandler returns a HTTP handler for API endpoints.
func MakeClientsHandler(svc users.Service, mux *bone.Mux, logger logger.Logger) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	mux.Post("/users", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("register_user"))(registrationEndpoint(svc)),
		decodeCreateUserReq,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/users/:id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("view_user"))(viewUserEndpoint(svc)),
		decodeViewUser,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/users/profile", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("view_profile"))(viewProfileEndpoint(svc)),
		decodeViewProfile,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/users", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("list_users"))(listClientsEndpoint(svc)),
		decodeListUsers,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/users/:groupID/members", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("list_members"))(listMembersEndpoint(svc)),
		decodeListMembersRequest,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/users/:id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_user_name_and_metadata"))(updateUserEndpoint(svc)),
		decodeUpdateUser,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/users/:id/tags", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_user_tags"))(updateUserTagsEndpoint(svc)),
		decodeUpdateUserTags,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/users/:id/identity", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_user_identity"))(updateUserIdentityEndpoint(svc)),
		decodeUpdateUserCredentials,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/password/reset-request", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("res-req"))(passwordResetRequestEndpoint(svc)),
		decodePasswordResetRequest,
		api.EncodeResponse,
		opts...,
	))

	mux.Put("/password/reset", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("reset"))(passwordResetEndpoint(svc)),
		decodePasswordReset,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/users/:id/password", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_user_password"))(changePasswordEndpoint(svc)),
		decodeUpdateUserCredentials,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/users/:id/owner", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_user_owner"))(updateUserOwnerEndpoint(svc)),
		decodeUpdateUserOwner,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/users/tokens", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("login"))(loginEndpoint(svc)),
		decodeCredentials,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/users/:id/enable", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("enable_user"))(enableUserEndpoint(svc)),
		decodeChangeClientStatus,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/users/:id/disable", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("disable_user"))(disableUserEndpoint(svc)),
		decodeChangeClientStatus,
		api.EncodeResponse,
		opts...,
	))

	mux.GetFunc("/health", mainflux.Health("users"))
	mux.Handle("/metrics", promhttp.Handler())
}

func decodeViewUser(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewUserReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}

	return req, nil
}

func decodeViewProfile(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewUserReq{token: apiutil.ExtractBearerToken(r)}

	return req, nil
}

func decodeListUsers(_ context.Context, r *http.Request) (interface{}, error) {
	var sid string
	s, err := apiutil.ReadStringQuery(r, api.StatusKey, api.DefClientStatus)
	if err != nil {
		return nil, err
	}
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, err
	}
	m, err := apiutil.ReadMetadataQuery(r, api.MetadataKey, nil)
	if err != nil {
		return nil, err
	}

	n, err := apiutil.ReadStringQuery(r, api.NameKey, "")
	if err != nil {
		return nil, err
	}
	t, err := apiutil.ReadStringQuery(r, api.TagKey, "")
	if err != nil {
		return nil, err
	}
	oid, err := apiutil.ReadStringQuery(r, api.OwnerKey, "")
	if err != nil {
		return nil, err
	}
	visibility, err := apiutil.ReadStringQuery(r, api.VisibilityKey, api.MyVisibility)
	if err != nil {
		return nil, err
	}
	switch visibility {
	case api.MyVisibility:
		oid = api.MyVisibility
	case api.SharedVisibility:
		sid = api.MyVisibility
	case api.AllVisibility:
		sid = api.MyVisibility
		oid = api.MyVisibility
	}
	st, err := users.ToStatus(s)
	if err != nil {
		return nil, err
	}
	req := listUsersReq{
		token:    apiutil.ExtractBearerToken(r),
		status:   st,
		offset:   o,
		limit:    l,
		metadata: m,
		name:     n,
		tag:      t,
		sharedBy: sid,
		owner:    oid,
	}
	return req, nil
}

func decodeUpdateUser(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateUserReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateUserTags(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateUserTagsReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodePasswordResetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var req passwResetReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	req.Host = r.Header.Get("Referer")
	return req, nil
}

func decodePasswordReset(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var req resetTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateUserCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateUserCredentialsReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateUserOwner(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateUserOwnerReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := loginUserReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCreateUserReq(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	req := createUserReq{
		user:  user,
		token: apiutil.ExtractBearerToken(r),
	}

	return req, nil
}

func decodeChangeClientStatus(_ context.Context, r *http.Request) (interface{}, error) {
	req := changeUserStatusReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}

	return req, nil
}

func decodeListMembersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	s, err := apiutil.ReadStringQuery(r, api.StatusKey, api.DefClientStatus)
	if err != nil {
		return nil, err
	}
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, err
	}
	m, err := apiutil.ReadMetadataQuery(r, api.MetadataKey, nil)
	if err != nil {
		return nil, err
	}
	st, err := users.ToStatus(s)
	if err != nil {
		return nil, err
	}
	req := listMembersReq{
		token: apiutil.ExtractBearerToken(r),
		Page: users.Page{
			Status:   st,
			Offset:   o,
			Limit:    l,
			Metadata: m,
		},
		groupID: bone.GetValue(r, "groupID"),
	}
	return req, nil
}
