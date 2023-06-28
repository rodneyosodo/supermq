// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/users/policies"
)

const (
	streamID  = "mainflux.users"
	streamLen = 1000
)

var _ policies.Service = (*eventStore)(nil)

type eventStore struct {
	svc    policies.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around policy service that sends
// events to event store.
func NewEventStoreMiddleware(svc policies.Service, client *redis.Client) policies.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) Authorize(ctx context.Context, ar policies.AccessRequest) error {
	if err := es.svc.Authorize(ctx, ar); err != nil {
		return err
	}

	event := authorizeEvent{
		ar,
	}
	values, err := event.Encode()
	if err != nil {
		return err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return err
	}

	return nil
}

func (es eventStore) AddPolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.AddPolicy(ctx, token, policy); err != nil {
		return err
	}

	event := policyEvent{
		policy, policyAdd,
	}
	values, err := event.Encode()
	if err != nil {
		return err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return err
	}

	return nil
}

func (es eventStore) UpdatePolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.UpdatePolicy(ctx, token, policy); err != nil {
		return err
	}

	event := policyEvent{
		policy, policyUpdate,
	}
	values, err := event.Encode()
	if err != nil {
		return err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return err
	}

	return nil
}

func (es eventStore) ListPolicies(ctx context.Context, token string, page policies.Page) (policies.PolicyPage, error) {
	pp, err := es.svc.ListPolicies(ctx, token, page)
	if err != nil {
		return pp, err
	}

	event := listPoliciesEvent{
		page,
	}
	values, err := event.Encode()
	if err != nil {
		return pp, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return pp, err
	}

	return pp, nil
}

func (es eventStore) DeletePolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.DeletePolicy(ctx, token, policy); err != nil {
		return err
	}

	event := policyEvent{
		policy, policyDelete,
	}
	values, err := event.Encode()
	if err != nil {
		return err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return err
	}

	return nil
}
