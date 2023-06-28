// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/mainflux/mainflux/users/jwt"
)

const (
	streamID  = "mainflux.users"
	streamLen = 1000
)

var _ clients.Service = (*eventStore)(nil)

type eventStore struct {
	svc    clients.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around users service that sends
// events to event store.
func NewEventStoreMiddleware(svc clients.Service, client *redis.Client) clients.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) RegisterClient(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.RegisterClient(ctx, token, user)
	if err != nil {
		return user, err
	}

	event := createClientEvent{
		user,
	}
	values, err := event.Encode()
	if err != nil {
		return user, err
	}
	record := &redis.XAddArgs{
		Stream: streamID,
		MaxLen: streamLen,
		Values: values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) UpdateClient(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.UpdateClient(ctx, token, user)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "", user)
}

func (es eventStore) UpdateClientOwner(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientOwner(ctx, token, user)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "owner", user)
}

func (es eventStore) UpdateClientTags(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientTags(ctx, token, user)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "tags", user)
}

func (es eventStore) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "secret", user)
}

func (es eventStore) UpdateClientIdentity(ctx context.Context, token, id, identity string) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientIdentity(ctx, token, id, identity)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "identity", user)
}

func (es eventStore) update(ctx context.Context, operation string, user mfclients.Client) (mfclients.Client, error) {
	event := updateClientEvent{
		user, operation,
	}
	values, err := event.Encode()
	if err != nil {
		return user, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) ViewClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	user, err := es.svc.ViewClient(ctx, token, id)
	if err != nil {
		return user, err
	}
	event := viewClientEvent{
		user,
	}
	values, err := event.Encode()
	if err != nil {
		return user, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) ViewProfile(ctx context.Context, token string) (mfclients.Client, error) {
	user, err := es.svc.ViewProfile(ctx, token)
	if err != nil {
		return user, err
	}
	event := viewProfileEvent{
		user,
	}
	values, err := event.Encode()
	if err != nil {
		return user, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	cp, err := es.svc.ListClients(ctx, token, pm)
	if err != nil {
		return cp, err
	}
	event := listClientEvent{
		pm,
	}
	values, err := event.Encode()
	if err != nil {
		return cp, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return cp, err
	}

	return cp, nil
}

func (es eventStore) ListMembers(ctx context.Context, token, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	mp, err := es.svc.ListMembers(ctx, token, groupID, pm)
	if err != nil {
		return mp, err
	}
	event := listClientByGroupEvent{
		pm, groupID,
	}
	values, err := event.Encode()
	if err != nil {
		return mp, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return mp, err
	}

	return mp, nil
}

func (es eventStore) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	user, err := es.svc.EnableClient(ctx, token, id)
	if err != nil {
		return user, err
	}

	return es.delete(ctx, user)
}

func (es eventStore) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	user, err := es.svc.DisableClient(ctx, token, id)
	if err != nil {
		return user, err
	}

	return es.delete(ctx, user)
}

func (es eventStore) delete(ctx context.Context, user mfclients.Client) (mfclients.Client, error) {
	event := removeClientEvent{
		id:        user.ID,
		updatedAt: user.UpdatedAt,
		updatedBy: user.UpdatedBy,
		status:    user.Status.String(),
	}
	values, err := event.Encode()
	if err != nil {
		return user, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) Identify(ctx context.Context, token string) (string, error) {
	userID, err := es.svc.Identify(ctx, token)
	if err != nil {
		return userID, err
	}
	event := identifyClientEvent{
		userID: userID,
	}
	values, err := event.Encode()
	if err != nil {
		return userID, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return userID, err
	}

	return userID, nil
}

func (es eventStore) GenerateResetToken(ctx context.Context, email, host string) error {
	if err := es.svc.GenerateResetToken(ctx, email, host); err != nil {
		return err
	}
	event := generateResetTokenEvent{
		email: email,
		host:  host,
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

func (es eventStore) IssueToken(ctx context.Context, identity, secret string) (jwt.Token, error) {
	token, err := es.svc.IssueToken(ctx, identity, secret)
	if err != nil {
		return token, err
	}
	event := issueTokenEvent{
		identity: identity,
	}
	values, err := event.Encode()
	if err != nil {
		return token, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return token, err
	}

	return token, nil
}

func (es eventStore) RefreshToken(ctx context.Context, refreshToken string) (jwt.Token, error) {
	token, err := es.svc.RefreshToken(ctx, refreshToken)
	if err != nil {
		return token, err
	}
	event := refreshTokenEvent{}
	values, err := event.Encode()
	if err != nil {
		return token, err
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       values,
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return token, err
	}

	return token, nil
}

func (es eventStore) ResetSecret(ctx context.Context, resetToken, secret string) error {
	if err := es.svc.ResetSecret(ctx, resetToken, secret); err != nil {
		return err
	}
	event := resetSecretEvent{}
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

func (es eventStore) SendPasswordReset(ctx context.Context, host, email, user, token string) error {
	if err := es.svc.SendPasswordReset(ctx, host, email, user, token); err != nil {
		return err
	}
	event := sendPasswordResetEvent{
		host:  host,
		email: email,
		user:  user,
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
