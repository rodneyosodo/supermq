// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	// CTJSON represents JSON content type.
	CTJSON ContentType = "application/json"

	// CTJSONSenML represents JSON SenML content type.
	CTJSONSenML ContentType = "application/senml+json"

	// CTBinary represents binary content type.
	CTBinary ContentType = "application/octet-stream"

	// EnabledStatus represents enable status for a client
	EnabledStatus = "enabled"

	// DisabledStatus represents disabled status for a client
	DisabledStatus = "disabled"
)

// ContentType represents all possible content types.
type ContentType string

var _ SDK = (*mfSDK)(nil)

var (
	// ErrFailedCreation indicates that entity creation failed.
	ErrFailedCreation = errors.New("failed to create entity in the db")

	// ErrFailedList indicates that entities list failed.
	ErrFailedList = errors.New("failed to list entities")

	// ErrFailedUpdate indicates that entity update failed.
	ErrFailedUpdate = errors.New("failed to update entity")

	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.New("failed to fetch entity")

	// ErrFailedRemoval indicates that entity removal failed.
	ErrFailedRemoval = errors.New("failed to remove entity")

	// ErrFailedEnable indicates that client enable failed.
	ErrFailedEnable = errors.New("failed to enable client")

	// ErrFailedDisable indicates that client disable failed.
	ErrFailedDisable = errors.New("failed to disable client")
)

type PageMetadata struct {
	Total        uint64   `json:"total"`
	Offset       uint64   `json:"offset"`
	Limit        uint64   `json:"limit"`
	Level        uint64   `json:"level,omitempty"`
	Email        string   `json:"email,omitempty"`
	Name         string   `json:"name,omitempty"`
	Type         string   `json:"type,omitempty"`
	Disconnected bool     `json:"disconnected,omitempty"`
	Metadata     Metadata `json:"metadata,omitempty"`
	Status       string   `json:"status,omitempty"`
	Action       string   `json:"action,omitempty"`
	Subject      string   `json:"subject,omitempty"`
	Object       string   `json:"object,omitempty"`
	Tag          string   `json:"tag,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	SharedBy     string   `json:"shared_by,omitempty"`
	Visibility   string   `json:"visibility,omitempty"`
	OwnerID      string   `json:"owner_id,omitempty"`
	Topic        string   `json:"topic,omitempty"`
	Contact      string   `json:"contact,omitempty"`
	State        string   `json:"state,omitempty"`
}

// Credentials represent client credentials: it contains
// "identity" which can be a username, email, generated name;
// and "secret" which can be a password or access token.
type Credentials struct {
	Identity string `json:"identity,omitempty"` // username or generated login ID
	Secret   string `json:"secret,omitempty"`   // password or token
}

// SDK contains Mainflux API.
type SDK interface {
	// CreateUser registers mainflux user.
	//
	// example:
	//  user := sdk.User{
	//    Name:	 "John Doe",
	//    Credentials: sdk.Credentials{
	//      Identity: "john.doe@example",
	//      Secret:   "12345678",
	//    },
	//  }
	//  user, _ := sdk.CreateUser(user)
	//  fmt.Println(user)
	CreateUser(user User, token string) (User, errors.SDKError)

	// User returns user object by id.
	//
	// example:
	//  user, _ := sdk.User("1", "token")
	//  fmt.Println(user)
	User(id, token string) (User, errors.SDKError)

	// Users returns list of users.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//		Name:   "John Doe",
	//	}
	//	users, _ := sdk.Users(pm, "token")
	//	fmt.Println(users)
	Users(pm PageMetadata, token string) (UsersPage, errors.SDKError)

	// Members retrieves everything that is assigned to a group identified by groupID.
	//
	// example:
	//	pm := sdk.PageMetadata{
	//		Offset: 0,
	//		Limit:  10,
	//	}
	//	members, _ := sdk.Members("1", pm, "token")
	//	fmt.Println(members)
	Members(groupID string, meta PageMetadata, token string) (MembersPage, errors.SDKError)

	// UserProfile returns user logged in.
	//
	// example:
	//  user, _ := sdk.UserProfile("token")
	//  fmt.Println(user)
	UserProfile(token string) (User, errors.SDKError)

	// UpdateUser updates existing user.
	//
	// example:
	//  user := sdk.User{
	//    ID:   "1",
	//    Name: "John Doe",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	// user, _ := sdk.UpdateUser(user, "token")
	// fmt.Println(user)
	UpdateUser(user User, token string) (User, errors.SDKError)

	// UpdateUserTags updates the user's tags.
	//
	// example:
	//  user := sdk.User{
	//    ID:   "1",
	//    Tags: []string{"tag1", "tag2"},
	//  }
	// user, _ := sdk.UpdateUserTags(user, "token")
	// fmt.Println(user)
	UpdateUserTags(user User, token string) (User, errors.SDKError)

	// UpdateUserIdentity updates the user's identity
	//
	// example:
	//  user := sdk.User{
	//    ID:   "1",
	//    Credentials: sdk.Credentials{
	//      Identity: "john.doe@example",
	//    },
	//  }
	// user, _ := sdk.UpdateUserIdentity(user, "token")
	// fmt.Println(user)
	UpdateUserIdentity(user User, token string) (User, errors.SDKError)

	// UpdateUserOwner updates the user's owner.
	//
	// example:
	//  user := sdk.User{
	//    ID:   "1",
	//    Owner: "2",
	//  }
	// user, _ := sdk.UpdateUserOwner(user, "token")
	// fmt.Println(user)
	UpdateUserOwner(user User, token string) (User, errors.SDKError)

	// UpdatePassword updates user password.
	//
	// example:
	//  user, _ := sdk.UpdatePassword("oldPass", "newPass", "token")
	//  fmt.Println(user)
	UpdatePassword(oldPass, newPass, token string) (User, errors.SDKError)

	// EnableUser changes the status of the user to enabled.
	//
	// example:
	//  user, _ := sdk.EnableUser("1", "token")
	//  fmt.Println(user)
	EnableUser(id, token string) (User, errors.SDKError)

	// DisableUser changes the status of the user to disabled.
	//
	// example:
	//  user, _ := sdk.DisableUser("1", "token")
	//  fmt.Println(user)
	DisableUser(id, token string) (User, errors.SDKError)

	// CreateToken receives credentials and returns user token.
	//
	// example:
	//  user := sdk.User{
	//    Credentials: sdk.Credentials{
	//      Identity: "john.doe@example",
	//      Secret:   "12345678",
	//    },
	//  }
	//  token, _ := sdk.CreateToken(user)
	//  fmt.Println(token)
	CreateToken(user User) (Token, errors.SDKError)

	// RefreshToken receives credentials and returns user token.
	//
	// example:
	//  token, _ := sdk.RefreshToken("refresh_token")
	//  fmt.Println(token)
	RefreshToken(token string) (Token, errors.SDKError)

	// CreateThing registers new thing and returns its id.
	//
	// example:
	//  thing := sdk.Thing{
	//    Name: "My Thing",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  thing, _ := sdk.CreateThing(thing, "token")
	//  fmt.Println(thing)
	CreateThing(thing Thing, token string) (Thing, errors.SDKError)

	// CreateThings registers new things and returns their ids.
	//
	// example:
	//  things := []sdk.Thing{
	//    {
	//      Name: "My Thing 1",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//    {
	//      Name: "My Thing 2",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//  }
	//  things, _ := sdk.CreateThings(things, "token")
	//  fmt.Println(things)
	CreateThings(things []Thing, token string) ([]Thing, errors.SDKError)

	// Filters things and returns a page result.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Thing",
	//  }
	//  things, _ := sdk.Things(pm, "token")
	//  fmt.Println(things)
	Things(pm PageMetadata, token string) (ThingsPage, errors.SDKError)

	// ThingsByChannel returns page of things that are connected or not connected
	// to specified channel.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Thing",
	//  }
	//  things, _ := sdk.ThingsByChannel("1", pm, "token")
	//  fmt.Println(things)
	ThingsByChannel(chanID string, pm PageMetadata, token string) (ThingsPage, errors.SDKError)

	// Thing returns thing object by id.
	//
	// example:
	//  thing, _ := sdk.Thing("1", "token")
	//  fmt.Println(thing)
	Thing(id, token string) (Thing, errors.SDKError)

	// UpdateThing updates existing thing.
	//
	// example:
	//  thing := sdk.Thing{
	//    ID:   "1",
	//    Name: "My Thing",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  thing, _ := sdk.UpdateThing(thing, "token")
	//  fmt.Println(thing)
	UpdateThing(thing Thing, token string) (Thing, errors.SDKError)

	// UpdateThingTags updates the client's tags.
	//
	// example:
	//  thing := sdk.Thing{
	//    ID:   "1",
	//    Tags: []string{"tag1", "tag2"},
	//  }
	//  thing, _ := sdk.UpdateThingTags(thing, "token")
	//  fmt.Println(thing)
	UpdateThingTags(thing Thing, token string) (Thing, errors.SDKError)

	// UpdateThingSecret updates the client's secret
	//
	// example:
	//  thing, err := sdk.UpdateThingSecret("1", "newSecret", "token")
	//  fmt.Println(thing)
	UpdateThingSecret(id, secret, token string) (Thing, errors.SDKError)

	// UpdateThingOwner updates the client's owner.
	//
	// example:
	//  thing := sdk.Thing{
	//    ID:    "1",
	//    Owner: "1",
	//  }
	//  thing, _ := sdk.UpdateThingOwner(thing, "token")
	//  fmt.Println(thing)
	UpdateThingOwner(thing Thing, token string) (Thing, errors.SDKError)

	// EnableThing changes client status to enabled.
	//
	// example:
	//  thing, _ := sdk.EnableThing("1", "token")
	//  fmt.Println(thing)
	EnableThing(id, token string) (Thing, errors.SDKError)

	// DisableThing changes client status to disabled - soft delete.
	//
	// example:
	//  thing, _ := sdk.DisableThing("1", "token")
	//  fmt.Println(thing)
	DisableThing(id, token string) (Thing, errors.SDKError)

	// IdentifyThing validates thing's key and returns its ID
	//
	// example:
	//  id, _ := sdk.IdentifyThing("key")
	//  fmt.Println(id)
	IdentifyThing(key string) (string, errors.SDKError)

	// CreateGroup creates new group and returns its id.
	//
	// example:
	//  group := sdk.Group{
	//    Name: "My Group",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  group, _ := sdk.CreateGroup(group, "token")
	//  fmt.Println(group)
	CreateGroup(group Group, token string) (Group, errors.SDKError)

	// Memberships
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Memberships("1", pm, "token")
	//  fmt.Println(groups)
	Memberships(clientID string, pm PageMetadata, token string) (MembershipsPage, errors.SDKError)

	// Groups returns page of groups.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Groups(pm, "token")
	//  fmt.Println(groups)
	Groups(pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Parents returns page of users groups.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Parents("1", pm, "token")
	//  fmt.Println(groups)
	Parents(id string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Children returns page of users groups.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Group",
	//  }
	//  groups, _ := sdk.Children("1", pm, "token")
	//  fmt.Println(groups)
	Children(id string, pm PageMetadata, token string) (GroupsPage, errors.SDKError)

	// Group returns users group object by id.
	//
	// example:
	//  group, _ := sdk.Group("1", "token")
	//  fmt.Println(group)
	Group(id, token string) (Group, errors.SDKError)

	// UpdateGroup updates existing group.
	//
	// example:
	//  group := sdk.Group{
	//    ID:   "1",
	//    Name: "My Group",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  group, _ := sdk.UpdateGroup(group, "token")
	//  fmt.Println(group)
	UpdateGroup(group Group, token string) (Group, errors.SDKError)

	// EnableGroup changes group status to enabled.
	//
	// example:
	//  group, _ := sdk.EnableGroup("1", "token")
	//  fmt.Println(group)
	EnableGroup(id, token string) (Group, errors.SDKError)

	// DisableGroup changes group status to disabled - soft delete.
	//
	// example:
	//  group, _ := sdk.DisableGroup("1", "token")
	//  fmt.Println(group)
	DisableGroup(id, token string) (Group, errors.SDKError)

	// CreateChannel creates new channel and returns its id.
	//
	// example:
	//  channel := sdk.Channel{
	//    Name: "My Channel",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  channel, _ := sdk.CreateChannel(channel, "token")
	//  fmt.Println(channel)
	CreateChannel(channel Channel, token string) (Channel, errors.SDKError)

	// CreateChannels registers new channels and returns their ids.
	//
	// example:
	//  channels := []sdk.Channel{
	//    {
	//      Name: "My Channel 1",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//    {
	//      Name: "My Channel 2",
	//      Metadata: sdk.Metadata{
	//        "key": "value",
	//      },
	//    },
	//  }
	//  channels, _ := sdk.CreateChannels(channels, "token")
	//  fmt.Println(channels)
	CreateChannels(channels []Channel, token string) ([]Channel, errors.SDKError)

	// Channels returns page of channels.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Channel",
	//  }
	//  channels, _ := sdk.Channels(pm, "token")
	//  fmt.Println(channels)
	Channels(pm PageMetadata, token string) (ChannelsPage, errors.SDKError)

	// ChannelsByThing returns page of channels that are connected or not connected
	// to specified thing.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Name:   "My Channel",
	//  }
	//  channels, _ := sdk.ChannelsByThing("1", pm, "token")
	//  fmt.Println(channels)
	ChannelsByThing(thingID string, pm PageMetadata, token string) (ChannelsPage, errors.SDKError)

	// Channel returns channel data by id.
	//
	// example:
	//  channel, _ := sdk.Channel("1", "token")
	//  fmt.Println(channel)
	Channel(id, token string) (Channel, errors.SDKError)

	// UpdateChannel updates existing channel.
	//
	// example:
	//  channel := sdk.Channel{
	//    ID:   "1",
	//    Name: "My Channel",
	//    Metadata: sdk.Metadata{
	//      "key": "value",
	//    },
	//  }
	//  channel, _ := sdk.UpdateChannel(channel, "token")
	//  fmt.Println(channel)
	UpdateChannel(channel Channel, token string) (Channel, errors.SDKError)

	// EnableChannel changes channel status to enabled.
	//
	// example:
	//  channel, _ := sdk.EnableChannel("1", "token")
	//  fmt.Println(channel)
	EnableChannel(id, token string) (Channel, errors.SDKError)

	// DisableChannel changes channel status to disabled - soft delete.
	//
	// example:
	//  channel, _ := sdk.DisableChannel("1", "token")
	//  fmt.Println(channel)
	DisableChannel(id, token string) (Channel, errors.SDKError)

	// CreatePolicy creates a policy for the given subject, so that, after
	// CreatePolicy, `subject` has a `relation` on `object`. Returns a non-nil
	// error in case of failures.
	//
	// example:
	//  policy := sdk.Policy{
	//    Subject: "user:1",
	//    Object:  "group:1",
	//    Actions: []string{"g_add"},
	//  }
	//  err := sdk.CreatePolicy(policy, "token")
	//  fmt.Println(err)
	CreatePolicy(policy Policy, token string) errors.SDKError

	// DeletePolicy deletes policies.
	//
	// example:
	//  policy := sdk.Policy{
	//    Subject: "user:1",
	//    Object:  "group:1",
	//    Actions: []string{"g_add"},
	//  }
	//  err := sdk.DeletePolicy(policy, "token")
	//  fmt.Println(err)
	DeletePolicy(policy Policy, token string) errors.SDKError

	// UpdatePolicy updates policies based on the given policy structure.
	//
	// example:
	//  policy := sdk.Policy{
	//    Subject: "user:1",
	//    Object:  "group:1",
	//    Actions: []string{"g_add"},
	//  }
	//  err := sdk.UpdatePolicy(policy, "token")
	//  fmt.Println(err)
	UpdatePolicy(p Policy, token string) errors.SDKError

	// ListPolicies lists policies based on the given policy structure.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//    Subject: "user:1",
	//  }
	//  policies, _ := sdk.ListPolicies(pm, "token")
	//  fmt.Println(policies)
	ListPolicies(pm PageMetadata, token string) (PolicyPage, errors.SDKError)

	// Assign assigns member of member type (thing or user) to a group.
	//
	// example:
	//  err := sdk.Assign([]string{"g_add"}, "user:1", "group:1", "token")
	//  fmt.Println(err)
	Assign(memberType []string, memberID, groupID, token string) errors.SDKError

	// Unassign removes member from a group.
	//
	// example:
	//  err := sdk.Unassign([]string{"g_add"}, "user:1", "group:1", "token")
	//  fmt.Println(err)
	Unassign(memberType []string, groupID string, memberID string, token string) errors.SDKError

	// Connect bulk connects things to channels specified by id.
	//
	// example:
	//  conns := sdk.ConnectionIDs{
	//    ChannelIDs: []string{"1", "2"},
	//    ThingIDs:   []string{"1", "2"},
	//  }
	//  err := sdk.Connect(conns, "token")
	//  fmt.Println(err)
	Connect(conns ConnectionIDs, token string) errors.SDKError

	// Disconnect
	//
	// example:
	//  conns := sdk.ConnectionIDs{
	//    ChannelIDs: []string{"1", "2"},
	//    ThingIDs:   []string{"1", "2"},
	//  }
	//  err := sdk.Disconnect(conns, "token")
	//  fmt.Println(err)
	Disconnect(connIDs ConnectionIDs, token string) errors.SDKError

	// ConnectThing
	//
	// example:
	//  err := sdk.ConnectThing("thing:1", "channel:1", "token")
	//  fmt.Println(err)
	ConnectThing(thingID, chanID, token string) errors.SDKError

	// DisconnectThing disconnect thing from specified channel by id.
	//
	// example:
	//  err := sdk.DisconnectThing("thing:1", "channel:1", "token")
	//  fmt.Println(err)
	DisconnectThing(thingID, chanID, token string) errors.SDKError

	// SendMessage send message to specified channel.
	//
	// example:
	//  msg := sdk.Message{
	//    Channel: "1",
	//    Payload: []byte("payload"),
	//  }
	//  err := sdk.SendMessage(msg, "token")
	//  fmt.Println(err)
	SendMessage(chanID, msg, key string) errors.SDKError

	// ReadMessages read messages of specified channel.
	//
	// example:
	//  msgs, _ := sdk.ReadMessages("1", "token")
	//  fmt.Println(msgs)
	ReadMessages(chanID, token string) (MessagesPage, errors.SDKError)

	// SetContentType sets message content type.
	//
	// example:
	//  err := sdk.SetContentType("application/json")
	//  fmt.Println(err)
	SetContentType(ct ContentType) errors.SDKError

	// Health returns things service health check.
	//
	// example:
	//  health, _ := sdk.Health()
	//  fmt.Println(health)
	Health() (mainflux.HealthInfo, errors.SDKError)

	// AddBootstrap add bootstrap configuration
	//
	// example:
	//  cfg := sdk.BootstrapConfig{
	//    ThingID: "thingID",
	//    Name: "bootstrap",
	//    ExternalID: "externalID",
	//    ExternalKey: "externalKey",
	//    Channels: []string{"channel1", "channel2"},
	//  }
	//  id, _ := sdk.AddBootstrap(cfg, "token")
	//  fmt.Println(id)
	AddBootstrap(cfg BootstrapConfig, token string) (string, errors.SDKError)

	// View returns Thing Config with given ID belonging to the user identified by the given token.
	//
	// example:
	//  bootstrap, _ := sdk.ViewBootstrap("id", "token")
	//  fmt.Println(bootstrap)
	ViewBootstrap(id, token string) (BootstrapConfig, errors.SDKError)

	// Update updates editable fields of the provided Config.
	//
	// example:
	//  cfg := sdk.BootstrapConfig{
	//    ThingID: "thingID",
	//    Name: "bootstrap",
	//    ExternalID: "externalID",
	//    ExternalKey: "externalKey",
	//    Channels: []string{"channel1", "channel2"},
	//  }
	//  err := sdk.UpdateBootstrap(cfg, "token")
	//  fmt.Println(err)
	UpdateBootstrap(cfg BootstrapConfig, token string) errors.SDKError

	// Update bootstrap config certificates.
	//
	// example:
	//  err := sdk.UpdateBootstrapCerts("id", "clientCert", "clientKey", "ca", "token")
	//  fmt.Println(err)
	UpdateBootstrapCerts(id string, clientCert, clientKey, ca string, token string) errors.SDKError

	// UpdateBootstrapConnection updates connections performs update of the channel list corresponding Thing is connected to.
	//
	// example:
	//  err := sdk.UpdateBootstrapConnection("id", []string{"channel1", "channel2"}, "token")
	//  fmt.Println(err)
	UpdateBootstrapConnection(id string, channels []string, token string) errors.SDKError

	// Remove removes Config with specified token that belongs to the user identified by the given token.
	//
	// example:
	//  err := sdk.RemoveBootstrap("id", "token")
	//  fmt.Println(err)
	RemoveBootstrap(id, token string) errors.SDKError

	// Bootstrap returns Config to the Thing with provided external ID using external key.
	//
	// example:
	//  bootstrap, _ := sdk.Bootstrap("externalID", "externalKey")
	//  fmt.Println(bootstrap)
	Bootstrap(externalID, externalKey string) (BootstrapConfig, errors.SDKError)

	// BootstrapSecure retrieves a configuration with given external ID and encrypted external key.
	//
	// example:
	//  bootstrap, _ := sdk.BootstrapSecure("externalID", "externalKey")
	//  fmt.Println(bootstrap)
	BootstrapSecure(externalID, externalKey string) (BootstrapConfig, errors.SDKError)

	// Bootstraps retrieves a list of managed configs.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//  }
	//  bootstraps, _ := sdk.Bootstraps(pm, "token")
	//  fmt.Println(bootstraps)
	Bootstraps(pm PageMetadata, token string) (BoostrapsPage, errors.SDKError)

	// Whitelist updates Thing state Config with given ID belonging to the user identified by the given token.
	//
	// example:
	//  cfg := sdk.BootstrapConfig{
	//    ThingID: "thingID",
	//    Name: "bootstrap",
	//    ExternalID: "externalID",
	//    ExternalKey: "externalKey",
	//    Channels: []string{"channel1", "channel2"},
	//  }
	//  err := sdk.Whitelist(cfg, "token")
	//  fmt.Println(err)
	Whitelist(cfg BootstrapConfig, token string) errors.SDKError

	// IssueCert issues a certificate for a thing required for mTLS.
	//
	// example:
	//  cert, _ := sdk.IssueCert("thingID", "valid", "token")
	//  fmt.Println(cert)
	IssueCert(thingID, valid, token string) (Cert, errors.SDKError)

	// ViewCert returns a certificate given certificate ID
	//
	// example:
	//  cert, _ := sdk.ViewCert("id", "token")
	//  fmt.Println(cert)
	ViewCert(certID, token string) (Cert, errors.SDKError)

	// ViewCertByThing retrieves a list of certificates' serial IDs for a given thing ID.
	//
	// example:
	//  cserial, _ := sdk.ViewCertByThing("thingID", "token")
	//  fmt.Println(cserial)
	ViewCertByThing(thingID, token string) (CertSerials, errors.SDKError)

	// RevokeCert revokes certificate for thing with thingID
	//
	// example:
	//  tm, _ := sdk.RevokeCert("thingID", "token")
	//  fmt.Println(tm)
	RevokeCert(thingID, token string) (time.Time, errors.SDKError)

	// CreateSubscription creates a new subscription
	//
	// example:
	//  subscription, _ := sdk.CreateSubscription("topic", "contact", "token")
	//  fmt.Println(subscription)
	CreateSubscription(topic, contact, token string) (string, errors.SDKError)

	// ListSubscriptions list subscriptions given list parameters.
	//
	// example:
	//  pm := sdk.PageMetadata{
	//    Offset: 0,
	//    Limit:  10,
	//  }
	//  subscriptions, _ := sdk.ListSubscriptions(pm, "token")
	//  fmt.Println(subscriptions)
	ListSubscriptions(pm PageMetadata, token string) (SubscriptionPage, errors.SDKError)

	// ViewSubscription retrieves a subscription with the provided id.
	//
	// example:
	//  subscription, _ := sdk.ViewSubscription("id", "token")
	//  fmt.Println(subscription)
	ViewSubscription(id, token string) (Subscription, errors.SDKError)

	// DeleteSubscription removes a subscription with the provided id.
	//
	// example:
	//  err := sdk.DeleteSubscription("id", "token")
	//  fmt.Println(err)
	DeleteSubscription(id, token string) errors.SDKError
}

type mfSDK struct {
	bootstrapURL   string
	certsURL       string
	httpAdapterURL string
	readerURL      string
	thingsURL      string
	usersURL       string

	msgContentType ContentType
	client         *http.Client
}

// Config contains sdk configuration parameters.
type Config struct {
	BootstrapURL   string
	CertsURL       string
	HTTPAdapterURL string
	ReaderURL      string
	ThingsURL      string
	UsersURL       string

	MsgContentType  ContentType
	TLSVerification bool
}

// NewSDK returns new mainflux SDK instance.
func NewSDK(conf Config) SDK {
	return &mfSDK{
		bootstrapURL:   conf.BootstrapURL,
		certsURL:       conf.CertsURL,
		httpAdapterURL: conf.HTTPAdapterURL,
		readerURL:      conf.ReaderURL,
		thingsURL:      conf.ThingsURL,
		usersURL:       conf.UsersURL,

		msgContentType: conf.MsgContentType,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !conf.TLSVerification,
				},
			},
		},
	}
}

// processRequest creates and send a new HTTP request, and checks for errors in the HTTP response.
// It then returns the response headers, the response body, and the associated error(s) (if any).
func (sdk mfSDK) processRequest(method, url, token, contentType string, data []byte, expectedRespCodes ...int) (http.Header, []byte, errors.SDKError) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}

	if token != "" {
		if !strings.Contains(token, apiutil.ThingPrefix) {
			token = apiutil.BearerPrefix + token
		}
		req.Header.Set("Authorization", token)
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := sdk.client.Do(req)
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}
	defer resp.Body.Close()

	sdkerr := errors.CheckError(resp, expectedRespCodes...)
	if sdkerr != nil {
		return make(http.Header), []byte{}, sdkerr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}

	return resp.Header, body, nil
}

func (sdk mfSDK) withQueryParams(baseURL, endpoint string, pm PageMetadata) (string, error) {
	q, err := pm.query()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s?%s", baseURL, endpoint, q), nil
}

func (pm PageMetadata) query() (string, error) {
	q := url.Values{}
	if pm.Offset != 0 {
		q.Add("offset", strconv.FormatUint(pm.Offset, 10))
	}
	if pm.Limit != 0 {
		q.Add("limit", strconv.FormatUint(pm.Limit, 10))
	}
	if pm.Total != 0 {
		q.Add("total", strconv.FormatUint(pm.Total, 10))
	}
	if pm.Level != 0 {
		q.Add("level", strconv.FormatUint(pm.Level, 10))
	}
	if pm.Email != "" {
		q.Add("email", pm.Email)
	}
	if pm.Name != "" {
		q.Add("name", pm.Name)
	}
	if pm.Type != "" {
		q.Add("type", pm.Type)
	}
	if pm.Visibility != "" {
		q.Add("visibility", pm.Visibility)
	}
	if pm.Status != "" {
		q.Add("status", pm.Status)
	}
	if pm.Metadata != nil {
		md, err := json.Marshal(pm.Metadata)
		if err != nil {
			return "", errors.NewSDKError(err)
		}
		q.Add("metadata", string(md))
	}
	return q.Encode(), nil
}
