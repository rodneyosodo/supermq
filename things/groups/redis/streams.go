// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	mfgroups "github.com/mainflux/mainflux/pkg/groups"
	"github.com/mainflux/mainflux/things/groups"
)

const (
	streamID                       = "mainflux.things"
	streamLen                      = 1000
	checkUnpublishedEventsInterval = 1 * time.Minute
)

var _ groups.Service = (*eventStore)(nil)

type eventStore struct {
	svc               groups.Service
	client            *redis.Client
	unpublishedEvents []*redis.XAddArgs
	mu                sync.Mutex
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc groups.Service, client *redis.Client) groups.Service {
	es := &eventStore{
		svc:    svc,
		client: client,
	}

	go es.startPublishingRoutine(ctx)

	return es
}

func (es *eventStore) CreateGroups(ctx context.Context, token string, groups ...mfgroups.Group) ([]mfgroups.Group, error) {
	gs, err := es.svc.CreateGroups(ctx, token, groups...)
	if err != nil {
		return gs, err
	}

	for _, group := range gs {
		event := createGroupEvent{
			group,
		}
		if err := es.publish(ctx, event); err != nil {
			return gs, err
		}
	}
	return gs, nil
}

func (es *eventStore) UpdateGroup(ctx context.Context, token string, group mfgroups.Group) (mfgroups.Group, error) {
	group, err := es.svc.UpdateGroup(ctx, token, group)
	if err != nil {
		return mfgroups.Group{}, err
	}

	event := updateGroupEvent{
		group,
	}
	if err := es.publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
}

func (es *eventStore) ViewGroup(ctx context.Context, token, id string) (mfgroups.Group, error) {
	group, err := es.svc.ViewGroup(ctx, token, id)
	if err != nil {
		return mfgroups.Group{}, err
	}
	event := viewGroupEvent{
		group,
	}
	if err := es.publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
}

func (es *eventStore) ListGroups(ctx context.Context, token string, pm mfgroups.GroupsPage) (mfgroups.GroupsPage, error) {
	gp, err := es.svc.ListGroups(ctx, token, pm)
	if err != nil {
		return mfgroups.GroupsPage{}, err
	}
	event := listGroupEvent{
		pm,
	}
	if err := es.publish(ctx, event); err != nil {
		return gp, err
	}

	return gp, nil
}

func (es *eventStore) ListMemberships(ctx context.Context, token, clientID string, pm mfgroups.GroupsPage) (mfgroups.MembershipsPage, error) {
	mp, err := es.svc.ListMemberships(ctx, token, clientID, pm)
	if err != nil {
		return mfgroups.MembershipsPage{}, err
	}
	event := listGroupMembershipEvent{
		pm, clientID,
	}
	if err := es.publish(ctx, event); err != nil {
		return mp, err
	}

	return mp, nil
}

func (es *eventStore) EnableGroup(ctx context.Context, token, id string) (mfgroups.Group, error) {
	cli, err := es.svc.EnableGroup(ctx, token, id)
	if err != nil {
		return mfgroups.Group{}, err
	}

	return es.delete(ctx, cli)
}

func (es *eventStore) DisableGroup(ctx context.Context, token, id string) (mfgroups.Group, error) {
	cli, err := es.svc.DisableGroup(ctx, token, id)
	if err != nil {
		return mfgroups.Group{}, err
	}

	return es.delete(ctx, cli)
}

func (es *eventStore) delete(ctx context.Context, group mfgroups.Group) (mfgroups.Group, error) {
	event := removeGroupEvent{
		id:        group.ID,
		updatedAt: group.UpdatedAt,
		updatedBy: group.UpdatedBy,
		status:    group.Status.String(),
	}
	if err := es.publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
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
				es.mu.Lock()
				for i := len(es.unpublishedEvents) - 1; i >= 0; i-- {
					if err := es.client.XAdd(ctx, es.unpublishedEvents[i]).Err(); err == nil {
						es.unpublishedEvents = append(es.unpublishedEvents[:i], es.unpublishedEvents[i+1:]...)
					}
				}
				es.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}
