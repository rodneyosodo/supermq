// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sort"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var _ users.Repository = (*userRepositoryMock)(nil)

type userRepositoryMock struct {
	mu             sync.Mutex
	users          map[string]users.User
	usersByID      map[string]users.User
	usersByGroupID map[string]users.User
}

// NewUserRepository creates in-memory user repository
func NewUserRepository() users.Repository {
	return &userRepositoryMock{
		users:          make(map[string]users.User),
		usersByID:      make(map[string]users.User),
		usersByGroupID: make(map[string]users.User),
	}
}

func (urm *userRepositoryMock) Save(ctx context.Context, user users.User) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Credentials.Identity]; ok {
		return users.User{}, errors.ErrConflict
	}

	urm.users[user.Credentials.Identity] = user
	urm.usersByID[user.ID] = user
	return user, nil
}

func (urm *userRepositoryMock) Update(ctx context.Context, user users.User) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Credentials.Identity]; !ok {
		return users.User{}, errors.ErrNotFound
	}

	urm.users[user.Credentials.Identity] = user
	return user, nil
}

func (urm *userRepositoryMock) UpdateTags(ctx context.Context, user users.User) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Credentials.Identity]; !ok {
		return users.User{}, errors.ErrNotFound
	}

	urm.users[user.Credentials.Identity] = user
	return user, nil
}

func (urm *userRepositoryMock) UpdateIdentity(ctx context.Context, user users.User) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Credentials.Identity]; !ok {
		return users.User{}, errors.ErrNotFound
	}

	urm.users[user.Credentials.Identity] = user
	return user, nil
}

func (urm *userRepositoryMock) UpdateSecret(ctx context.Context, user users.User) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Credentials.Identity]; !ok {
		return users.User{}, errors.ErrNotFound
	}

	urm.users[user.Credentials.Identity] = user
	return user, nil
}

func (urm *userRepositoryMock) UpdateOwner(ctx context.Context, user users.User) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Credentials.Identity]; !ok {
		return users.User{}, errors.ErrNotFound
	}

	urm.users[user.Credentials.Identity] = user
	return user, nil
}

func (urm *userRepositoryMock) RetrieveByIdentity(ctx context.Context, identity string) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	val, ok := urm.users[identity]
	if !ok {
		return users.User{}, errors.ErrNotFound
	}

	return val, nil
}

func (urm *userRepositoryMock) RetrieveByID(ctx context.Context, id string) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	val, ok := urm.usersByID[id]
	if !ok {
		return users.User{}, errors.ErrNotFound
	}

	return val, nil
}

func (urm *userRepositoryMock) RetrieveAll(ctx context.Context, pm users.Page) (users.UsersPage, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	up := users.UsersPage{}
	i := uint64(0)

	if pm.Identity != "" {
		val, ok := urm.users[pm.Identity]
		if !ok {
			return users.UsersPage{}, errors.ErrNotFound
		}
		up.Offset = pm.Offset
		up.Limit = pm.Limit
		up.Total = uint64(i)
		up.Users = []users.User{val}
		return up, nil
	}

	if pm.Status == users.EnabledStatus || pm.Status == users.DisabledStatus {
		for _, u := range sortUsers(urm.users) {
			if i >= pm.Offset && i < (pm.Limit+pm.Offset) {
				if pm.Status == u.Status {
					up.Users = append(up.Users, u)
				}
			}
			i++
		}
		up.Offset = pm.Offset
		up.Limit = pm.Limit
		up.Total = uint64(i)
		return up, nil
	}
	for _, u := range sortUsers(urm.users) {
		if i >= pm.Offset && i < (pm.Limit+pm.Offset) {
			up.Users = append(up.Users, u)
		}
		i++
	}

	up.Offset = pm.Offset
	up.Limit = pm.Limit
	up.Total = uint64(i)

	return up, nil
}

func (urm *userRepositoryMock) Members(ctx context.Context, groupID string, pm users.Page) (users.MembersPage, error) {
	panic("unimplemented")
}

func (urm *userRepositoryMock) ChangeStatus(ctx context.Context, id string, status users.Status) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	user, ok := urm.usersByID[id]
	if !ok {
		return users.User{}, errors.ErrNotFound
	}
	user.Status = status
	urm.usersByID[id] = user
	urm.users[user.Credentials.Identity] = user
	return user, nil
}

func sortUsers(us map[string]users.User) []users.User {
	users := []users.User{}
	ids := make([]string, 0, len(us))
	for k := range us {
		ids = append(ids, k)
	}

	sort.Strings(ids)
	for _, id := range ids {
		users = append(users, us[id])
	}

	return users
}
