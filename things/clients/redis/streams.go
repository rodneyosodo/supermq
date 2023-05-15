// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/things/clients"
)

const (
	streamID  = "mainflux.things"
	streamLen = 1000
)

var _ clients.Service = (*eventStore)(nil)

type eventStore struct {
	svc    clients.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(svc clients.Service, client *redis.Client) clients.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) CreateThings(ctx context.Context, token string, thing ...mfclients.Client) ([]mfclients.Client, error) {
	sths, err := es.svc.CreateThings(ctx, token, thing...)
	if err != nil {
		return sths, err
	}

	for _, th := range sths {
		event := createClientEvent{
			th,
		}
		record := &redis.XAddArgs{
			Stream: streamID,
			MaxLen: streamLen,
			Values: event.Encode(),
		}
		if err := es.client.XAdd(ctx, record).Err(); err != nil {
			return sths, err
		}
	}
	return sths, nil
}

func (es eventStore) UpdateClient(ctx context.Context, token string, thing mfclients.Client) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClient(ctx, token, thing)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "", cli)
}

func (es eventStore) UpdateClientOwner(ctx context.Context, token string, thing mfclients.Client) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClientOwner(ctx, token, thing)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "owner", cli)
}

func (es eventStore) UpdateClientTags(ctx context.Context, token string, thing mfclients.Client) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClientTags(ctx, token, thing)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "tags", cli)
}

func (es eventStore) UpdateClientSecret(ctx context.Context, token, id, key string) (mfclients.Client, error) {
	cli, err := es.svc.UpdateClientSecret(ctx, token, id, key)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.update(ctx, "secret", cli)
}

func (es eventStore) update(ctx context.Context, operation string, cli mfclients.Client) (mfclients.Client, error) {
	event := updateClientEvent{
		cli, operation,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return mfclients.Client{}, err
	}

	return cli, nil
}

func (es eventStore) ShareClient(ctx context.Context, token, thingID string, actions, userIDs []string) error {
	if err := es.svc.ShareClient(ctx, token, thingID, actions, userIDs); err != nil {
		return err
	}
	event := shareClientEvent{
		thingID: thingID,
		actions: fmt.Sprintf("[%s]", strings.Join(actions, ",")),
		userIDs: fmt.Sprintf("[%s]", strings.Join(userIDs, ",")),
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return err
	}

	return nil
}

func (es eventStore) ViewClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	cli, err := es.svc.ViewClient(ctx, token, id)
	if err != nil {
		return mfclients.Client{}, err
	}
	event := viewClientEvent{
		cli,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return mfclients.Client{}, err
	}

	return cli, nil
}

func (es eventStore) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	cp, err := es.svc.ListClients(ctx, token, pm)
	if err != nil {
		return mfclients.ClientsPage{}, err
	}
	event := listClientEvent{
		pm,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return mfclients.ClientsPage{}, err
	}

	return cp, nil
}

func (es eventStore) ListClientsByGroup(ctx context.Context, token, chID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	cp, err := es.svc.ListClientsByGroup(ctx, token, chID, pm)
	if err != nil {
		return mfclients.MembersPage{}, err
	}
	event := listClientByGroupEvent{
		pm, chID,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return mfclients.MembersPage{}, err
	}

	return cp, nil
}

func (es eventStore) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	cli, err := es.svc.EnableClient(ctx, token, id)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.delete(ctx, cli)
}

func (es eventStore) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	cli, err := es.svc.DisableClient(ctx, token, id)
	if err != nil {
		return mfclients.Client{}, err
	}

	return es.delete(ctx, cli)
}

func (es eventStore) delete(ctx context.Context, cli mfclients.Client) (mfclients.Client, error) {
	event := removeClientEvent{
		id:        cli.ID,
		updatedAt: cli.UpdatedAt,
		updatedBy: cli.UpdatedBy,
		status:    cli.Status.String(),
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return mfclients.Client{}, err
	}

	return cli, nil
}

func (es eventStore) Identify(ctx context.Context, key string) (string, error) {
	thingID, err := es.svc.Identify(ctx, key)
	if err != nil {
		return "", err
	}
	event := identifyClientEvent{
		thingID: thingID,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return "", err
	}

	return thingID, nil
}
