// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/things/policies"
)

const (
	streamID                       = "mainflux.things"
	streamLen                      = 1000
	checkUnpublishedEventsInterval = 1 * time.Minute
)

var _ policies.Service = (*eventStore)(nil)

type eventStore struct {
	svc               policies.Service
	client            *redis.Client
	unpublishedEvents []*redis.XAddArgs
}

// NewEventStoreMiddleware returns wrapper around policy service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc policies.Service, client *redis.Client) policies.Service {
	es := &eventStore{
		svc:    svc,
		client: client,
	}

	go es.startPublishingRoutine(ctx)

	return es
}

func (es *eventStore) Authorize(ctx context.Context, ar policies.AccessRequest) (policies.Policy, error) {
	id, err := es.svc.Authorize(ctx, ar)
	if err != nil {
		return policies.Policy{}, err
	}

	event := authorizeEvent{
		ar, ar.Entity,
	}
	if err := es.publish(ctx, event); err != nil {
		return id, err
	}

	return id, nil
}

func (es *eventStore) AddPolicy(ctx context.Context, token string, policy policies.Policy) (policies.Policy, error) {
	policy, err := es.svc.AddPolicy(ctx, token, policy)
	if err != nil {
		return policies.Policy{}, err
	}

	event := policyEvent{
		policy, policyAdd,
	}
	if err := es.publish(ctx, event); err != nil {
		return policy, err
	}

	return policy, nil
}

func (es *eventStore) UpdatePolicy(ctx context.Context, token string, policy policies.Policy) (policies.Policy, error) {
	policy, err := es.svc.UpdatePolicy(ctx, token, policy)
	if err != nil {
		return policies.Policy{}, err
	}

	event := policyEvent{
		policy, policyUpdate,
	}
	if err := es.publish(ctx, event); err != nil {
		return policy, err
	}

	return policy, nil
}

func (es *eventStore) ListPolicies(ctx context.Context, token string, page policies.Page) (policies.PolicyPage, error) {
	policypage, err := es.svc.ListPolicies(ctx, token, page)
	if err != nil {
		return policies.PolicyPage{}, err
	}

	event := listPoliciesEvent{
		page,
	}
	if err := es.publish(ctx, event); err != nil {
		return policypage, err
	}

	return policypage, nil
}

func (es *eventStore) DeletePolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.DeletePolicy(ctx, token, policy); err != nil {
		return err
	}

	event := policyEvent{
		policy, policyDelete,
	}

	return es.publish(ctx, event)
}

func (es *eventStore) checkRedisConnection(ctx context.Context) error {
	// A timeout is used to avoid blocking the main thread
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	return es.client.Ping(ctx).Err()
}

func (es *eventStore) publish(ctx context.Context, event event) error {
	values, err := event.Encode()
	if err != nil {
		return err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}

	if err := es.checkRedisConnection(ctx); err != nil {
		es.unpublishedEvents = append(es.unpublishedEvents, record)
		return nil
	}

	return es.client.XAdd(ctx, record).Err()
}

func (es *eventStore) startPublishingRoutine(ctx context.Context) {
	ticker := time.NewTicker(checkUnpublishedEventsInterval)
	for {
		select {
		case <-ticker.C:
			if err := es.checkRedisConnection(ctx); err == nil {
				for i := len(es.unpublishedEvents) - 1; i >= 0; i-- {
					if err := es.client.XAdd(ctx, es.unpublishedEvents[i]).Err(); err == nil {
						es.unpublishedEvents = append(es.unpublishedEvents[:i], es.unpublishedEvents[i+1:]...)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
