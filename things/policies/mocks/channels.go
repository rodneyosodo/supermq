// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"strings"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

const separator = ":"

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

func (ccm *cacheMock) Put(_ context.Context, policy policies.Policy, thingID string) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	key, value := kv(policy, thingID)
	ccm.policies[key] = value

	return nil
}

func (ccm *cacheMock) Get(_ context.Context, policy policies.Policy) (policies.Policy, string, error) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	key, _ := kv(policy, "")

	val := ccm.policies[key]
	if val == "" {
		return policies.Policy{}, "", errors.ErrNotFound
	}

	thingID := extractThingID(val)
	if thingID == "" {
		return policies.Policy{}, "", errors.ErrNotFound
	}

	policy.Actions = separateActions(val)

	return policy, thingID, nil
}

func (ccm *cacheMock) Remove(_ context.Context, policy policies.Policy) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	key, _ := kv(policy, "")

	delete(ccm.policies, key)

	return nil
}

// kv is used to create a key-value pair for caching.
// If thingID is not empty, it will be appended to the value.
func kv(p policies.Policy, thingID string) (string, string) {
	if thingID != "" {
		return p.Subject + separator + p.Object, strings.Join(p.Actions, separator) + separator + thingID
	}

	return p.Subject + separator + p.Object, strings.Join(p.Actions, separator)
}

// separateActions is used to separate the actions from the cache values.
func separateActions(actions string) []string {
	return strings.Split(actions, separator)
}

// extractThingID is used to extract the thingID from the cache values.
func extractThingID(actions string) string {
	var lastIdx = strings.LastIndex(actions, separator)

	thingID := actions[lastIdx+1:]
	// check if the thingID is a valid UUID
	if len(thingID) != 36 {
		return ""
	}

	return thingID
}
