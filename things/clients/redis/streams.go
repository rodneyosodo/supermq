// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/things/clients"
)

const (
	streamID                       = "mainflux.things"
	streamLen                      = 1000
	checkUnpublishedEventsInterval = 1 * time.Minute
)

var _ clients.Service = (*eventStore)(nil)

type eventStore struct {
	svc               clients.Service
	client            *redis.Client
	unpublishedEvents []*redis.XAddArgs
	mu                sync.Mutex
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc clients.Service, client *redis.Client) clients.Service {
	es := &eventStore{
		svc:    svc,
		client: client,
	}

	go es.startPublishingRoutine(ctx)

	return es
}

func (es *eventStore) CreateThings(ctx context.Context, token string, thing ...mfclients.Client) ([]mfclients.Client, error) {
	sths, err := es.svc.CreateThings(ctx, token, thing...)
	if err != nil {
		return sths, err
	}

	for _, th := range sths {
		event := createClientEvent{
			th,
		}
		if err := es.publish(ctx, event); err != nil {
			return sths, err
		}
	}
	return sths, nil
}

func (es *eventStore) UpdateClient(ctx context.Context, token string, thing mfclients.Client) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClient(ctx, token, thing)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "", cli)
}

func (es *eventStore) UpdateClientOwner(ctx context.Context, token string, thing mfclients.Client) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClientOwner(ctx, token, thing)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "owner", cli)
}

func (es *eventStore) UpdateClientTags(ctx context.Context, token string, thing mfclients.Client) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClientTags(ctx, token, thing)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "tags", cli)
}

func (es *eventStore) UpdateClientSecret(ctx context.Context, token, id, key string) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClientSecret(ctx, token, id, key)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "secret", cli)
}

func (es *eventStore) update(ctx context.Context, operation string, cli mfclients.Client) (mfclients.Client, error) {
	event := updateClientEvent{
		cli, operation,
	}
	if err := es.publish(ctx, event); err != nil {
		return cli, err
	}

	return cli, nil
}

func (es *eventStore) ViewClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	cli, err := es.svc.ViewClient(ctx, token, id)
	if err != nil {
		return mfclients.Client{}, err
	}
	event := viewClientEvent{
		cli,
	}
	if err := es.publish(ctx, event); err != nil {
		return cli, err
	}

	return cli, nil
}

func (es *eventStore) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	cp, err := es.svc.ListClients(ctx, token, pm)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}
	event := listClientEvent{
		pm,
	}
	if err := es.publish(ctx, event); err != nil {
		return cp, err
	}

	return cp, nil
}

func (es *eventStore) ListClientsByGroup(ctx context.Context, token, chID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	cp, err := es.svc.ListClientsByGroup(ctx, token, chID, pm)
	if err != nil {
		return mfclients.MembersPage{}, err
	}
	event := listClientByGroupEvent{
		pm, chID,
	}
	if err := es.publish(ctx, event); err != nil {
		return cp, err
	}

	return cp, nil
}

func (es *eventStore) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	cli, err := es.svc.EnableClient(ctx, token, id)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.delete(ctx, cli)
}

func (es *eventStore) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	cli, err := es.svc.DisableClient(ctx, token, id)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.delete(ctx, cli)
}

func (es *eventStore) delete(ctx context.Context, cli mfclients.Client) (mfclients.Client, error) {
	event := removeClientEvent{
		id:        cli.ID,
		updatedAt: cli.UpdatedAt,
		updatedBy: cli.UpdatedBy,
		status:    cli.Status.String(),
	}
	if err := es.publish(ctx, event); err != nil {
		return cli, err
	}

	return cli, nil
}

func (es *eventStore) Identify(ctx context.Context, key string) (string, error) {
	thingID, err := es.svc.Identify(ctx, key)
	if err != nil {
		return "", err
	}
	event := identifyClientEvent{
		thingID: thingID,
	}

	if err := es.publish(ctx, event); err != nil {
		return thingID, err
	}
	return thingID, nil
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
