// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	authn "github.com/absmach/magistrala/pkg/authn"

	mock "github.com/stretchr/testify/mock"

	things "github.com/absmach/magistrala/things"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Authorize provides a mock function with given fields: ctx, req
func (_m *Service) Authorize(ctx context.Context, req things.AuthzReq) (string, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for Authorize")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, things.AuthzReq) (string, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, things.AuthzReq) string); ok {
		r0 = rf(ctx, req)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, things.AuthzReq) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateClients provides a mock function with given fields: ctx, session, client
func (_m *Service) CreateClients(ctx context.Context, session authn.Session, client ...things.Client) ([]things.Client, error) {
	_va := make([]interface{}, len(client))
	for _i := range client {
		_va[_i] = client[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, session)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateClients")
	}

	var r0 []things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, ...things.Client) ([]things.Client, error)); ok {
		return rf(ctx, session, client...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, ...things.Client) []things.Client); ok {
		r0 = rf(ctx, session, client...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]things.Client)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, ...things.Client) error); ok {
		r1 = rf(ctx, session, client...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, session, id
func (_m *Service) Delete(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Disable provides a mock function with given fields: ctx, session, id
func (_m *Service) Disable(ctx context.Context, session authn.Session, id string) (things.Client, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for Disable")
	}

	var r0 things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (things.Client, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) things.Client); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(things.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Enable provides a mock function with given fields: ctx, session, id
func (_m *Service) Enable(ctx context.Context, session authn.Session, id string) (things.Client, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for Enable")
	}

	var r0 things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (things.Client, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) things.Client); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(things.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Identify provides a mock function with given fields: ctx, key
func (_m *Service) Identify(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Identify")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListClients provides a mock function with given fields: ctx, session, reqUserID, pm
func (_m *Service) ListClients(ctx context.Context, session authn.Session, reqUserID string, pm things.Page) (things.ClientsPage, error) {
	ret := _m.Called(ctx, session, reqUserID, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListClients")
	}

	var r0 things.ClientsPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, things.Page) (things.ClientsPage, error)); ok {
		return rf(ctx, session, reqUserID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, things.Page) things.ClientsPage); ok {
		r0 = rf(ctx, session, reqUserID, pm)
	} else {
		r0 = ret.Get(0).(things.ClientsPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, things.Page) error); ok {
		r1 = rf(ctx, session, reqUserID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListClientsByGroup provides a mock function with given fields: ctx, session, groupID, pm
func (_m *Service) ListClientsByGroup(ctx context.Context, session authn.Session, groupID string, pm things.Page) (things.MembersPage, error) {
	ret := _m.Called(ctx, session, groupID, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListClientsByGroup")
	}

	var r0 things.MembersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, things.Page) (things.MembersPage, error)); ok {
		return rf(ctx, session, groupID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, things.Page) things.MembersPage); ok {
		r0 = rf(ctx, session, groupID, pm)
	} else {
		r0 = ret.Get(0).(things.MembersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, things.Page) error); ok {
		r1 = rf(ctx, session, groupID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Share provides a mock function with given fields: ctx, session, id, relation, userids
func (_m *Service) Share(ctx context.Context, session authn.Session, id string, relation string, userids ...string) error {
	_va := make([]interface{}, len(userids))
	for _i := range userids {
		_va[_i] = userids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, session, id, relation)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Share")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, ...string) error); ok {
		r0 = rf(ctx, session, id, relation, userids...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unshare provides a mock function with given fields: ctx, session, id, relation, userids
func (_m *Service) Unshare(ctx context.Context, session authn.Session, id string, relation string, userids ...string) error {
	_va := make([]interface{}, len(userids))
	for _i := range userids {
		_va[_i] = userids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, session, id, relation)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Unshare")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, ...string) error); ok {
		r0 = rf(ctx, session, id, relation, userids...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, session, client
func (_m *Service) Update(ctx context.Context, session authn.Session, client things.Client) (things.Client, error) {
	ret := _m.Called(ctx, session, client)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, things.Client) (things.Client, error)); ok {
		return rf(ctx, session, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, things.Client) things.Client); ok {
		r0 = rf(ctx, session, client)
	} else {
		r0 = ret.Get(0).(things.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, things.Client) error); ok {
		r1 = rf(ctx, session, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateSecret provides a mock function with given fields: ctx, session, id, key
func (_m *Service) UpdateSecret(ctx context.Context, session authn.Session, id string, key string) (things.Client, error) {
	ret := _m.Called(ctx, session, id, key)

	if len(ret) == 0 {
		panic("no return value specified for UpdateSecret")
	}

	var r0 things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) (things.Client, error)); ok {
		return rf(ctx, session, id, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) things.Client); ok {
		r0 = rf(ctx, session, id, key)
	} else {
		r0 = ret.Get(0).(things.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, id, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateTags provides a mock function with given fields: ctx, session, client
func (_m *Service) UpdateTags(ctx context.Context, session authn.Session, client things.Client) (things.Client, error) {
	ret := _m.Called(ctx, session, client)

	if len(ret) == 0 {
		panic("no return value specified for UpdateTags")
	}

	var r0 things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, things.Client) (things.Client, error)); ok {
		return rf(ctx, session, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, things.Client) things.Client); ok {
		r0 = rf(ctx, session, client)
	} else {
		r0 = ret.Get(0).(things.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, things.Client) error); ok {
		r1 = rf(ctx, session, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// View provides a mock function with given fields: ctx, session, id
func (_m *Service) View(ctx context.Context, session authn.Session, id string) (things.Client, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for View")
	}

	var r0 things.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (things.Client, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) things.Client); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(things.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewPerms provides a mock function with given fields: ctx, session, id
func (_m *Service) ViewPerms(ctx context.Context, session authn.Session, id string) ([]string, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for ViewPerms")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) ([]string, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) []string); ok {
		r0 = rf(ctx, session, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
