// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

type cacheMock struct {
	mu       sync.Mutex
	policies map[string]string
}

// NewCache returns mock cache instance.
func NewCache() policies.Cache {
	return &cacheMock{
		policies: make(map[string]string),
	}
}

func (ccm *cacheMock) Put(_ context.Context, key, value string) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	ccm.policies[key] = value
	return nil
}

func (ccm *cacheMock) Get(_ context.Context, key string) (string, error) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()
	actions := ccm.policies[key]

	if actions != "" {
		return actions, nil
	}

	return "", errors.ErrNotFound
}

func (ccm *cacheMock) Remove(_ context.Context, key string) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	delete(ccm.policies, key)
	return nil
}
