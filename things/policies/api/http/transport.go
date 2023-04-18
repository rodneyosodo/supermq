package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/policies"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakePolicyHandler(csvc clients.Service, psvc policies.Service, mux *bone.Mux, logger logger.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Post("/connect", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("connect"))(connectThingsEndpoint(psvc)),
		decodeConnectList,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/disconnect", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("disconnect"))(disconnectThingsEndpoint(psvc)),
		decodeConnectList,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/channels/:chanId/things/:thingId", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("connect_thing"))(connectEndpoint(psvc)),
		decodeConnectThing,
		api.EncodeResponse,
		opts...,
	))

	mux.Delete("/channels/:chanId/things/:thingId", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("disconnect_thing"))(disconnectEndpoint(psvc)),
		decodeDisconnectThing,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/identify", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("identify"))(identifyEndpoint(csvc)),
		decodeIdentify,
		api.EncodeResponse,
		opts...,
	))

	mux.Put("/identify", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_policy"))(updatePolicyEndpoint(psvc)),
		decodeUpdatePolicy,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/identify", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("list_policies"))(listPoliciesEndpoint(psvc)),
		decodeListPolicies,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/identify/channels/:chanId/access-by-key", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("authorize_by_key"))(authorizeByKeyEndpoint(psvc)),
		decodeCanAccessByKey,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/identify/channels/:chanId/access-by-id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("authorize"))(authorizeEndpoint(psvc)),
		decodeCanAccessByID,
		api.EncodeResponse,
		opts...,
	))
	return mux

}

func decodeConnectThing(_ context.Context, r *http.Request) (interface{}, error) {
	req := createPolicyReq{
		token:    apiutil.ExtractBearerToken(r),
		GroupID:  bone.GetValue(r, "chanId"),
		ClientID: bone.GetValue(r, "thingId"),
	}
	if r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}

	return req, nil
}

func decodeDisconnectThing(_ context.Context, r *http.Request) (interface{}, error) {
	req := createPolicyReq{
		token:    apiutil.ExtractBearerToken(r),
		GroupID:  bone.GetValue(r, "chanId"),
		ClientID: bone.GetValue(r, "thingId"),
	}
	if r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}

	return req, nil
}

func decodeConnectList(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := createPoliciesReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeIdentify(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	req := identifyReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCanAccessByKey(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	req := authorizeReq{
		GroupID: bone.GetValue(r, "chanId"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCanAccessByID(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := authorizeReq{
		GroupID: bone.GetValue(r, "chanId"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdatePolicy(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := policyReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeListPolicies(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, err
	}
	c, err := apiutil.ReadStringQuery(r, api.ClientKey, "")
	if err != nil {
		return nil, err
	}
	g, err := apiutil.ReadStringQuery(r, api.GroupKey, "")
	if err != nil {
		return nil, err
	}
	a, err := apiutil.ReadStringQuery(r, api.ActionKey, "")
	if err != nil {
		return nil, err
	}
	oid, err := apiutil.ReadStringQuery(r, api.OwnerKey, "")
	if err != nil {
		return nil, err
	}

	req := listPoliciesReq{
		token:  apiutil.ExtractBearerToken(r),
		offset: o,
		limit:  l,
		client: c,
		group:  g,
		action: a,
		owner:  oid,
	}

	return req, nil
}