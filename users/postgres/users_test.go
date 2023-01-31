package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	cpostgres "github.com/mainflux/mainflux/users/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ultravioletrs/clients/groups"
	gpostgres "github.com/ultravioletrs/clients/groups/postgres"
	"github.com/ultravioletrs/clients/policies"
	ppostgres "github.com/ultravioletrs/clients/policies/postgres"
)

const (
	maxNameSize = 254
)

var (
	idProvider   = uuid.New()
	invalidName  = strings.Repeat("m", maxNameSize+10)
	password     = "$tr0ngPassw0rd"
	userIdentity = "user-identity@example.com"
	userName     = "user name"
	wrongName    = "wrong-name"
	wrongID      = "wrong-id"
)

func TestClientsSave(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	uid := testsutil.GenerateUUID(t, idProvider)

	cases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "add new user successfully",
			user: users.User{
				ID:   uid,
				Name: userName,
				Credentials: users.Credentials{
					Identity: userIdentity,
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "add new user with an owner",
			user: users.User{
				ID:    testsutil.GenerateUUID(t, idProvider),
				Owner: uid,
				Name:  userName,
				Credentials: users.Credentials{
					Identity: "withowner-user@example.com",
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "add user with duplicate user identity",
			user: users.User{
				ID:   testsutil.GenerateUUID(t, idProvider),
				Name: userName,
				Credentials: users.Credentials{
					Identity: userIdentity,
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: errors.ErrConflict,
		},
		{
			desc: "add user with invalid user id",
			user: users.User{
				ID:   invalidName,
				Name: userName,
				Credentials: users.Credentials{
					Identity: "invalidid-user@example.com",
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add user with invalid user name",
			user: users.User{
				ID:   testsutil.GenerateUUID(t, idProvider),
				Name: invalidName,
				Credentials: users.Credentials{
					Identity: "invalidname-user@example.com",
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add user with invalid user owner",
			user: users.User{
				ID:    testsutil.GenerateUUID(t, idProvider),
				Owner: invalidName,
				Credentials: users.Credentials{
					Identity: "invalidowner-user@example.com",
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add user with invalid user identity",
			user: users.User{
				ID:   testsutil.GenerateUUID(t, idProvider),
				Name: userName,
				Credentials: users.Credentials{
					Identity: invalidName,
					Secret:   password,
				},
				Metadata: users.Metadata{},
				Status:   users.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add user with a missing user identity",
			user: users.User{
				ID: testsutil.GenerateUUID(t, idProvider),
				Credentials: users.Credentials{
					Identity: "",
					Secret:   password,
				},
				Metadata: users.Metadata{},
			},
			err: nil,
		},
		{
			desc: "add user with a missing user secret",
			user: users.User{
				ID: testsutil.GenerateUUID(t, idProvider),
				Credentials: users.Credentials{
					Identity: "missing-user-secret@example.com",
					Secret:   "",
				},
				Metadata: users.Metadata{},
			},
			err: nil,
		},
	}
	for _, tc := range cases {
		rUser, err := repo.Save(context.Background(), tc.user)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		rUser.Credentials.Secret = tc.user.Credentials.Secret
		if err == nil {
			assert.Equal(t, tc.user, rUser, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.user, rUser))
		}
	}
}

func TestClientsRetrieveByID(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: userName,
		Credentials: users.Credentials{
			Identity: userIdentity,
			Secret:   password,
		},
		Status: users.EnabledStatus,
	}

	user, err := repo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		ID  string
		err error
	}{
		"retrieve existing user":     {user.ID, nil},
		"retrieve non-existing user": {wrongID, errors.ErrNotFound},
	}

	for desc, tc := range cases {
		rUser, err := repo.RetrieveByID(context.Background(), tc.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		if err == nil {
			assert.Equal(t, user.ID, rUser.ID, fmt.Sprintf("retrieve user by ID : user ID : expected %s got %s\n", user.ID, rUser.ID))
			assert.Equal(t, user.Name, rUser.Name, fmt.Sprintf("retrieve user by ID : user Name : expected %s got %s\n", user.Name, rUser.Name))
			assert.Equal(t, user.Credentials.Identity, rUser.Credentials.Identity, fmt.Sprintf("retrieve user by ID : user Identity : expected %s got %s\n", user.Credentials.Identity, rUser.Credentials.Identity))
			assert.Equal(t, user.Status, rUser.Status, fmt.Sprintf("retrieve user by ID : user Status : expected %d got %d\n", user.Status, rUser.Status))
		}
	}
}

func TestClientsRetrieveByIdentity(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: userName,
		Credentials: users.Credentials{
			Identity: userIdentity,
			Secret:   password,
		},
		Status: users.EnabledStatus,
	}

	_, err := repo.Save(context.Background(), user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		identity string
		err      error
	}{
		"retrieve existing user":     {userIdentity, nil},
		"retrieve non-existing user": {wrongID, errors.ErrNotFound},
	}

	for desc, tc := range cases {
		_, err := repo.RetrieveByIdentity(context.Background(), tc.identity)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestClientsRetrieveAll(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)
	grepo := gpostgres.NewGroupRepo(database)
	prepo := ppostgres.NewPolicyRepo(database)

	var nUsers = uint64(200)
	var ownerID string

	meta := users.Metadata{
		"admin": "true",
	}
	wrongMeta := users.Metadata{
		"admin": "false",
	}
	var expectedUsers = []users.User{}

	var sharedGroup = groups.Group{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "shared-group",
	}
	_, err := grepo.Save(context.Background(), sharedGroup)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	for i := uint64(0); i < nUsers; i++ {
		identity := fmt.Sprintf("TestRetrieveAll%d@example.com", i)
		user := users.User{
			ID:   testsutil.GenerateUUID(t, idProvider),
			Name: identity,
			Credentials: users.Credentials{
				Identity: identity,
				Secret:   password,
			},
			Metadata: users.Metadata{},
			Status:   users.EnabledStatus,
		}
		if i == 1 {
			ownerID = user.ID
		}
		if i%10 == 0 {
			user.Owner = ownerID
			user.Metadata = meta
			user.Tags = []string{"Test"}
		}
		if i%50 == 0 {
			user.Status = users.DisabledStatus
		}
		_, err := repo.Save(context.Background(), user)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		user.Credentials.Secret = ""
		expectedUsers = append(expectedUsers, user)
		var policy = policies.Policy{
			Subject: user.ID,
			Object:  sharedGroup.ID,
			Actions: []string{"c_list"},
		}
		err = prepo.Save(context.Background(), policy)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	}

	cases := map[string]struct {
		size     uint64
		pm       users.Page
		response []users.User
	}{
		"retrieve all users empty page": {
			pm:       users.Page{},
			response: []users.User{},
			size:     0,
		},
		"retrieve all users": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Status: users.AllStatus,
			},
			response: expectedUsers,
			size:     200,
		},
		"retrieve all users with limit": {
			pm: users.Page{
				Offset: 0,
				Limit:  50,
				Status: users.AllStatus,
			},
			response: expectedUsers[0:50],
			size:     50,
		},
		"retrieve all users with offset": {
			pm: users.Page{
				Offset: 50,
				Limit:  nUsers,
				Status: users.AllStatus,
			},
			response: expectedUsers[50:200],
			size:     150,
		},
		"retrieve all users with limit and offset": {
			pm: users.Page{
				Offset: 50,
				Limit:  50,
				Status: users.AllStatus,
			},
			response: expectedUsers[50:100],
			size:     50,
		},
		"retrieve all users with limit and offset not full": {
			pm: users.Page{
				Offset: 170,
				Limit:  50,
				Status: users.AllStatus,
			},
			response: expectedUsers[170:200],
			size:     30,
		},
		"retrieve all users by metadata": {
			pm: users.Page{
				Offset:   0,
				Limit:    nUsers,
				Total:    nUsers,
				Metadata: meta,
				Status:   users.AllStatus,
			},
			response: []users.User{expectedUsers[0], expectedUsers[10], expectedUsers[20], expectedUsers[30], expectedUsers[40], expectedUsers[50], expectedUsers[60],
				expectedUsers[70], expectedUsers[80], expectedUsers[90], expectedUsers[100], expectedUsers[110], expectedUsers[120], expectedUsers[130],
				expectedUsers[140], expectedUsers[150], expectedUsers[160], expectedUsers[170], expectedUsers[180], expectedUsers[190],
			},
			size: 20,
		},
		"retrieve users by wrong metadata": {
			pm: users.Page{
				Offset:   0,
				Limit:    nUsers,
				Total:    nUsers,
				Metadata: wrongMeta,
				Status:   users.AllStatus,
			},
			response: []users.User{},
			size:     0,
		},
		"retrieve all users by name": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Name:   "TestRetrieveAll3@example.com",
				Status: users.AllStatus,
			},
			response: []users.User{expectedUsers[3]},
			size:     1,
		},
		"retrieve users by wrong name": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Name:   wrongName,
				Status: users.AllStatus,
			},
			response: []users.User{},
			size:     0,
		},
		"retrieve all users by owner": {
			pm: users.Page{
				Offset:  0,
				Limit:   nUsers,
				Total:   nUsers,
				OwnerID: ownerID,
				Status:  users.AllStatus,
			},
			response: []users.User{expectedUsers[10], expectedUsers[20], expectedUsers[30], expectedUsers[40], expectedUsers[50], expectedUsers[60],
				expectedUsers[70], expectedUsers[80], expectedUsers[90], expectedUsers[100], expectedUsers[110], expectedUsers[120], expectedUsers[130],
				expectedUsers[140], expectedUsers[150], expectedUsers[160], expectedUsers[170], expectedUsers[180], expectedUsers[190],
			},
			size: 19,
		},
		"retrieve users by wrong owner": {
			pm: users.Page{
				Offset:  0,
				Limit:   nUsers,
				Total:   nUsers,
				OwnerID: wrongID,
				Status:  users.AllStatus,
			},
			response: []users.User{},
			size:     0,
		},
		"retrieve all users by disabled status": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Status: users.DisabledStatus,
			},
			response: []users.User{expectedUsers[0], expectedUsers[50], expectedUsers[100], expectedUsers[150]},
			size:     4,
		},
		"retrieve all users by combined status": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Status: users.AllStatus,
			},
			response: expectedUsers,
			size:     200,
		},
		"retrieve users by the wrong status": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Status: 10,
			},
			response: []users.User{},
			size:     0,
		},
		"retrieve all users by tags": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Tag:    "Test",
				Status: users.AllStatus,
			},
			response: []users.User{expectedUsers[0], expectedUsers[10], expectedUsers[20], expectedUsers[30], expectedUsers[40], expectedUsers[50], expectedUsers[60],
				expectedUsers[70], expectedUsers[80], expectedUsers[90], expectedUsers[100], expectedUsers[110], expectedUsers[120], expectedUsers[130],
				expectedUsers[140], expectedUsers[150], expectedUsers[160], expectedUsers[170], expectedUsers[180], expectedUsers[190],
			},
			size: 20,
		},
		"retrieve users by wrong tags": {
			pm: users.Page{
				Offset: 0,
				Limit:  nUsers,
				Total:  nUsers,
				Tag:    "wrongTags",
				Status: users.AllStatus,
			},
			response: []users.User{},
			size:     0,
		},
		"retrieve all users by sharedby": {
			pm: users.Page{
				Offset:   0,
				Limit:    nUsers,
				Total:    nUsers,
				SharedBy: expectedUsers[0].ID,
				Status:   users.AllStatus,
				Action:   "c_list",
			},
			response: []users.User{expectedUsers[10], expectedUsers[20], expectedUsers[30], expectedUsers[40], expectedUsers[50], expectedUsers[60],
				expectedUsers[70], expectedUsers[80], expectedUsers[90], expectedUsers[100], expectedUsers[110], expectedUsers[120], expectedUsers[130],
				expectedUsers[140], expectedUsers[150], expectedUsers[160], expectedUsers[170], expectedUsers[180], expectedUsers[190],
			},
			size: 19,
		},
	}
	for desc, tc := range cases {
		page, err := repo.RetrieveAll(context.Background(), tc.pm)
		size := uint64(len(page.Users))
		assert.ElementsMatch(t, page.Users, tc.response, fmt.Sprintf("%s: expected %v got %v\n", desc, tc.response, page.Users))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestClientsUpdateMetadata(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user1 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-user",
		Credentials: users.Credentials{
			Identity: "user1-update@example.com",
			Secret:   password,
		},
		Metadata: users.Metadata{
			"name": "enabled-user",
		},
		Tags:   []string{"enabled", "tag1"},
		Status: users.EnabledStatus,
	}

	user2 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-user",
		Credentials: users.Credentials{
			Identity: "user2-update@example.com",
			Secret:   password,
		},
		Metadata: users.Metadata{
			"name": "disabled-user",
		},
		Tags:   []string{"disabled", "tag1"},
		Status: users.DisabledStatus,
	}

	user1, err := repo.Save(context.Background(), user1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new user with metadata: expected %v got %s\n", nil, err))
	user2, err = repo.Save(context.Background(), user2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled user: expected %v got %s\n", nil, err))

	ucases := []struct {
		desc   string
		update string
		user   users.User
		err    error
	}{
		{
			desc:   "update metadata for enabled user",
			update: "metadata",
			user: users.User{
				ID: user1.ID,
				Metadata: users.Metadata{
					"update": "metadata",
				},
			},
			err: nil,
		},
		{
			desc:   "update metadata for disabled user",
			update: "metadata",
			user: users.User{
				ID: user2.ID,
				Metadata: users.Metadata{
					"update": "metadata",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name for enabled user",
			update: "name",
			user: users.User{
				ID:   user1.ID,
				Name: "updated name",
			},
			err: nil,
		},
		{
			desc:   "update name for disabled user",
			update: "name",
			user: users.User{
				ID:   user2.ID,
				Name: "updated name",
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name and metadata for enabled user",
			update: "both",
			user: users.User{
				ID:   user1.ID,
				Name: "updated name and metadata",
				Metadata: users.Metadata{
					"update": "name and metadata",
				},
			},
			err: nil,
		},
		{
			desc:   "update name and metadata for a disabled user",
			update: "both",
			user: users.User{
				ID:   user2.ID,
				Name: "updated name and metadata",
				Metadata: users.Metadata{
					"update": "name and metadata",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update metadata for invalid user",
			update: "metadata",
			user: users.User{
				ID: wrongID,
				Metadata: users.Metadata{
					"update": "metadata",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name for invalid user",
			update: "name",
			user: users.User{
				ID:   wrongID,
				Name: "updated name",
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name and metadata for invalid user",
			update: "both",
			user: users.User{
				ID:   user2.ID,
				Name: "updated name and metadata",
				Metadata: users.Metadata{
					"update": "name and metadata",
				},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.Update(context.Background(), tc.user)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			if tc.user.Name != "" {
				assert.Equal(t, expected.Name, tc.user.Name, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, expected.Name, tc.user.Name))
			}
			if tc.user.Metadata != nil {
				assert.Equal(t, expected.Metadata, tc.user.Metadata, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, expected.Metadata, tc.user.Metadata))
			}

		}
	}
}

func TestClientsUpdateTags(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user1 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-user-with-tags",
		Credentials: users.Credentials{
			Identity: "user1-update-tags@example.com",
			Secret:   password,
		},
		Tags:   []string{"test", "enabled"},
		Status: users.EnabledStatus,
	}
	user2 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-user-with-tags",
		Credentials: users.Credentials{
			Identity: "user2-update-tags@example.com",
			Secret:   password,
		},
		Tags:   []string{"test", "disabled"},
		Status: users.DisabledStatus,
	}

	user1, err := repo.Save(context.Background(), user1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new user with tags: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user1.ID, user1.ID, fmt.Sprintf("add new user with tags: expected %v got %s\n", nil, err))
	}
	user2, err = repo.Save(context.Background(), user2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled user with tags: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user2.ID, user2.ID, fmt.Sprintf("add new disabled user with tags: expected %v got %s\n", nil, err))
	}
	ucases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "update tags for enabled user",
			user: users.User{
				ID:   user1.ID,
				Tags: []string{"updated"},
			},
			err: nil,
		},
		{
			desc: "update tags for disabled user",
			user: users.User{
				ID:   user2.ID,
				Tags: []string{"updated"},
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update tags for invalid user",
			user: users.User{
				ID:   wrongID,
				Tags: []string{"updated"},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.UpdateTags(context.Background(), tc.user)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.user.Tags, expected.Tags, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.user.Tags, expected.Tags))
		}
	}
}

func TestClientsUpdateSecret(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user1 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-user",
		Credentials: users.Credentials{
			Identity: "user1-update@example.com",
			Secret:   password,
		},
		Status: users.EnabledStatus,
	}
	user2 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-user",
		Credentials: users.Credentials{
			Identity: "user2-update@example.com",
			Secret:   password,
		},
		Status: users.DisabledStatus,
	}

	rUser1, err := repo.Save(context.Background(), user1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new user: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user1.ID, rUser1.ID, fmt.Sprintf("add new user: expected %v got %s\n", nil, err))
	}
	rUser2, err := repo.Save(context.Background(), user2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled user: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user2.ID, rUser2.ID, fmt.Sprintf("add new disabled user: expected %v got %s\n", nil, err))
	}

	ucases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "update secret for enabled user",
			user: users.User{
				ID: user1.ID,
				Credentials: users.Credentials{
					Identity: "user1-update@example.com",
					Secret:   "newpassword",
				},
			},
			err: nil,
		},
		{
			desc: "update secret for disabled user",
			user: users.User{
				ID: user2.ID,
				Credentials: users.Credentials{
					Identity: "user2-update@example.com",
					Secret:   "newpassword",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update secret for invalid user",
			user: users.User{
				ID: wrongID,
				Credentials: users.Credentials{
					Identity: "user3-update@example.com",
					Secret:   "newpassword",
				},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		_, err := repo.UpdateSecret(context.Background(), tc.user)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			c, err := repo.RetrieveByIdentity(context.Background(), tc.user.Credentials.Identity)
			require.Nil(t, err, fmt.Sprintf("retrieve user by id during update of secret unexpected error: %s", err))
			assert.Equal(t, tc.user.Credentials.Secret, c.Credentials.Secret, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.user.Credentials.Secret, c.Credentials.Secret))
		}
	}
}

func TestClientsUpdateIdentity(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user1 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-user",
		Credentials: users.Credentials{
			Identity: "user1-update@example.com",
			Secret:   password,
		},
		Status: users.EnabledStatus,
	}
	user2 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-user",
		Credentials: users.Credentials{
			Identity: "user2-update@example.com",
			Secret:   password,
		},
		Status: users.DisabledStatus,
	}

	rUser1, err := repo.Save(context.Background(), user1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new user: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user1.ID, rUser1.ID, fmt.Sprintf("add new user: expected %v got %s\n", nil, err))
	}
	rUser2, err := repo.Save(context.Background(), user2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled user: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user2.ID, rUser2.ID, fmt.Sprintf("add new disabled user: expected %v got %s\n", nil, err))
	}

	ucases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "update identity for enabled user",
			user: users.User{
				ID: user1.ID,
				Credentials: users.Credentials{
					Identity: "user1-updated@example.com",
				},
			},
			err: nil,
		},
		{
			desc: "update identity for disabled user",
			user: users.User{
				ID: user2.ID,
				Credentials: users.Credentials{
					Identity: "user2-updated@example.com",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update identity for invalid user",
			user: users.User{
				ID: wrongID,
				Credentials: users.Credentials{
					Identity: "user3-updated@example.com",
				},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.UpdateIdentity(context.Background(), tc.user)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.user.Credentials.Identity, expected.Credentials.Identity, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.user.Credentials.Identity, expected.Credentials.Identity))
		}
	}
}

func TestClientsUpdateOwner(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user1 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-user-with-owner",
		Credentials: users.Credentials{
			Identity: "user1-update-owner@example.com",
			Secret:   password,
		},
		Owner:  testsutil.GenerateUUID(t, idProvider),
		Status: users.EnabledStatus,
	}
	user2 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-user-with-owner",
		Credentials: users.Credentials{
			Identity: "user2-update-owner@example.com",
			Secret:   password,
		},
		Owner:  testsutil.GenerateUUID(t, idProvider),
		Status: users.DisabledStatus,
	}

	user1, err := repo.Save(context.Background(), user1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new user with owner: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user1.ID, user1.ID, fmt.Sprintf("add new user with owner: expected %v got %s\n", nil, err))
	}
	user2, err = repo.Save(context.Background(), user2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled user with owner: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, user2.ID, user2.ID, fmt.Sprintf("add new disabled user with owner: expected %v got %s\n", nil, err))
	}
	ucases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "update owner for enabled user",
			user: users.User{
				ID:    user1.ID,
				Owner: testsutil.GenerateUUID(t, idProvider),
			},
			err: nil,
		},
		{
			desc: "update owner for disabled user",
			user: users.User{
				ID:    user2.ID,
				Owner: testsutil.GenerateUUID(t, idProvider),
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update owner for invalid user",
			user: users.User{
				ID:    wrongID,
				Owner: testsutil.GenerateUUID(t, idProvider),
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.UpdateOwner(context.Background(), tc.user)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.user.Owner, expected.Owner, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.user.Owner, expected.Owner))
		}
	}
}

func TestClientsChangeStatus(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewUsersRepo(database)

	user1 := users.User{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-user",
		Credentials: users.Credentials{
			Identity: "user1-update@example.com",
			Secret:   password,
		},
		Status: users.EnabledStatus,
	}

	user1, err := repo.Save(context.Background(), user1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new user: expected %v got %s\n", nil, err))

	ucases := []struct {
		desc string
		user users.User
		err  error
	}{
		{
			desc: "change user status for an enabled user",
			user: users.User{
				ID:     user1.ID,
				Status: 0,
			},
			err: nil,
		},
		{
			desc: "change user status for a disabled user",
			user: users.User{
				ID:     user1.ID,
				Status: 1,
			},
			err: nil,
		},
		{
			desc: "change user status for non-existing user",
			user: users.User{
				ID:     "invalid",
				Status: 2,
			},
			err: errors.ErrNotFound,
		},
	}

	for _, tc := range ucases {
		expected, err := repo.ChangeStatus(context.Background(), tc.user.ID, tc.user.Status)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.user.Status, expected.Status, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.user.Status, expected.Status))
		}
	}
}
