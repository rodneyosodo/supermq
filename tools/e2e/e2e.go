// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"time"

	"github.com/gookit/color"
	"github.com/goombaio/namegenerator"
	"github.com/gorilla/websocket"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"golang.org/x/sync/errgroup"
)

const (
	defPass      = "12345678"
	defReaderURL = "http://localhost:8905"
	defWSPort    = "8186"
)

var (
	seed           = time.Now().UTC().UnixNano()
	namesgenerator = namegenerator.NewNameGenerator(seed)
	msg            = "[{'bn':'demo', 'bu':'V','bver':5, 'n':'voltage','u':'V','v':120}]"
)

// Config - test configuration
type Config struct {
	Host     string
	Num      uint64
	NumOfMsg uint64
	SSL      bool
	CA       string
	CAKey    string
	Prefix   string
}

// Test - function that does actual end to end testing
func Test(conf Config) {

	sdkConf := sdk.Config{
		ThingsURL:       fmt.Sprintf("http://%s", conf.Host),
		UsersURL:        fmt.Sprintf("http://%s", conf.Host),
		ReaderURL:       defReaderURL,
		HTTPAdapterURL:  fmt.Sprintf("http://%s/http", conf.Host),
		BootstrapURL:    fmt.Sprintf("http://%s", conf.Host),
		CertsURL:        fmt.Sprintf("http://%s", conf.Host),
		MsgContentType:  sdk.CTJSONSenML,
		TLSVerification: false,
	}

	s := sdk.NewSDK(sdkConf)

	/*
		- Create user
		- Create another user
		- Do CRUD on them

		- Create groups using hierachy
		- Do CRUD on them

		- Do CRUD on things
		- Do CRUD on channels

		- Connect thing to channel
		- Test messaging
	*/
	magenta := color.FgLightMagenta.Render

	token, owner, err := createUser(s, conf)
	if err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Printf("created user with token %s\n", magenta(token))

	users, err := createUsers(s, conf, token)
	if err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("created users of ids: ", magenta(getIDS(users)))

	groups, err := createGroups(s, conf, token)
	if err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("created groups of ids: ", magenta(getIDS(groups)))

	things, err := createThings(s, conf, token)
	if err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("created things of ids: ", magenta(getIDS(things)))

	channels, err := createChannels(s, conf, token)
	if err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("created channels of ids: ", magenta(getIDS(channels)))

	if err := createPolicies(s, conf, token, owner, users, groups, things, channels); err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("created policies for users, groups, things and channels")
	// List users, groups, things and channels
	if err := read(s, conf, token, users, groups, things, channels); err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("viewed users, groups, things and channels")

	// Update users, groups, things and channels
	if err := update(s, token, users, groups, things, channels); err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("updated users, groups, things and channels")

	// Send and Receive messages from channels
	if err := messaging(s, conf, token, things, channels); err != nil {
		color.Error.Println(err.Error())
		os.Exit(1)
	}
	color.Success.Println("sent and received messages from channels")
}

func createUser(s sdk.SDK, conf Config) (string, string, error) {
	user := sdk.User{
		Name: fmt.Sprintf("%s-%s", conf.Prefix, namesgenerator.Generate()),
		Credentials: sdk.Credentials{
			Identity: fmt.Sprintf("%s-%s@email.com", conf.Prefix, namesgenerator.Generate()),
			Secret:   defPass,
		},
		Status: "enabled",
	}

	pass := user.Credentials.Secret

	// Create new user
	user, err := s.CreateUser(user, "")
	if err != nil {
		return "", "", fmt.Errorf("unable to create user: %s", err.Error())
	}

	user.Credentials.Secret = pass
	// Login user
	token, err := s.CreateToken(user)
	if err != nil {
		return "", "", fmt.Errorf("unable to login user: %s", err.Error())
	}
	return token.AccessToken, user.ID, nil
}

func createUsers(s sdk.SDK, conf Config, token string) ([]sdk.User, error) {
	var err error
	users := []sdk.User{}

	for i := uint64(0); i < conf.Num; i++ {
		user := sdk.User{
			Name: fmt.Sprintf("%s-%s", conf.Prefix, namesgenerator.Generate()),
			Credentials: sdk.Credentials{
				Identity: fmt.Sprintf("%s-%s@email.com", conf.Prefix, namesgenerator.Generate()),
				Secret:   defPass},
			Status: "enabled",
		}

		user, err = s.CreateUser(user, token)
		if err != nil {
			return []sdk.User{}, fmt.Errorf("Failed to create the users: %s", err.Error())
		}
		users = append(users, user)

	}

	return users, nil
}

func createGroups(s sdk.SDK, conf Config, token string) ([]sdk.Group, error) {
	var err error
	groups := []sdk.Group{}

	parentID := ""
	for i := uint64(0); i < conf.Num; i++ {
		group := sdk.Group{
			Name:     fmt.Sprintf("%s-%s", conf.Prefix, namesgenerator.Generate()),
			ParentID: parentID,
			Status:   "enabled",
		}

		group, err = s.CreateGroup(group, token)
		if err != nil {
			return []sdk.Group{}, fmt.Errorf("Failed to create the group: %s", err.Error())
		}
		groups = append(groups, group)
		parentID = group.ID
	}

	return groups, nil
}

func createThings(s sdk.SDK, conf Config, token string) ([]sdk.Thing, error) {
	var err error
	things := make([]sdk.Thing, conf.Num)

	for i := uint64(0); i < conf.Num; i++ {
		things[i] = sdk.Thing{
			Name:   fmt.Sprintf("%s-%s", conf.Prefix, namesgenerator.Generate()),
			Status: "enabled",
		}
	}
	things, err = s.CreateThings(things, token)
	if err != nil {
		return []sdk.Thing{}, fmt.Errorf("Failed to create the things: %s", err.Error())
	}

	return things, nil
}

func createChannels(s sdk.SDK, conf Config, token string) ([]sdk.Channel, error) {
	var err error
	channels := make([]sdk.Channel, conf.Num)

	for i := uint64(0); i < conf.Num; i++ {
		channels[i] = sdk.Channel{
			Name:   fmt.Sprintf("%s-%s", conf.Prefix, namesgenerator.Generate()),
			Status: "enabled",
		}
	}

	channels, err = s.CreateChannels(channels, token)
	if err != nil {
		return []sdk.Channel{}, fmt.Errorf("Failed to create the chennels: %s", err.Error())
	}

	return channels, nil
}

func createPolicies(s sdk.SDK, conf Config, token, owner string, users []sdk.User, groups []sdk.Group, things []sdk.Thing, channels []sdk.Channel) error {
	for i := uint64(0); i < conf.Num; i++ {
		upolicy := sdk.Policy{
			Subject: owner,
			Object:  users[i].ID,
			Actions: []string{"c_delete", "c_update", "c_add", "c_list"},
		}
		gpolicy := sdk.Policy{
			Subject: owner,
			Object:  groups[i].ID,
			Actions: []string{"g_delete", "g_update", "g_add", "g_list"},
		}
		tpolicy := sdk.Policy{
			Subject: owner,
			Object:  things[i].ID,
			Actions: []string{"c_delete", "c_update", "c_add", "c_list"},
		}
		cpolicy := sdk.Policy{
			Subject: owner,
			Object:  channels[i].ID,
			Actions: []string{"g_delete", "g_update", "g_add", "g_list"},
		}
		if err := s.AddPolicy(upolicy, token); err != nil {
			return err
		}
		if err := s.AddPolicy(gpolicy, token); err != nil {
			return err
		}
		if err := s.AddPolicy(tpolicy, token); err != nil {
			return err
		}
		if err := s.AddPolicy(cpolicy, token); err != nil {
			return err
		}
	}
	return nil
}

func read(s sdk.SDK, conf Config, token string, users []sdk.User, groups []sdk.Group, things []sdk.Thing, channels []sdk.Channel) error {
	for _, user := range users {
		if _, err := s.User(user.ID, token); err != nil {
			return err
		}
	}
	up, err := s.Users(sdk.PageMetadata{}, token)
	if err != nil {
		return err
	}
	if up.Total != uint64(conf.Num) {
		return fmt.Errorf("returned users %d not equal to create users %d", up.Total, conf.Num)
	}
	for _, group := range groups {
		if _, err := s.Group(group.ID, token); err != nil {
			return err
		}
	}
	gp, err := s.Groups(sdk.PageMetadata{}, token)
	if err != nil {
		return err
	}
	if gp.Total != uint64(conf.Num) {
		return fmt.Errorf("returned groups %d not equal to create groups %d", gp.Total, conf.Num)
	}
	for _, thing := range things {
		if _, err := s.Thing(thing.ID, token); err != nil {
			return err
		}
	}
	tp, err := s.Things(sdk.PageMetadata{}, token)
	if err != nil {
		return err
	}
	if tp.Total != uint64(conf.Num) {
		return fmt.Errorf("returned things %d not equal to create things %d", tp.Total, conf.Num)
	}
	for _, channel := range channels {
		if _, err := s.Channel(channel.ID, token); err != nil {
			return err
		}
	}
	cp, err := s.Channels(sdk.PageMetadata{}, token)
	if err != nil {
		return err
	}
	if cp.Total != uint64(conf.Num) {
		return fmt.Errorf("returned channels %d not equal to create channels %d", cp.Total, conf.Num)
	}
	return nil
}

func update(s sdk.SDK, token string, users []sdk.User, groups []sdk.Group, things []sdk.Thing, channels []sdk.Channel) error {
	for _, user := range users {
		user.Name = namesgenerator.Generate()
		user.Metadata = sdk.Metadata{"Update": namesgenerator.Generate()}
		rUser, err := s.UpdateUser(user, token)
		if err != nil {
			return fmt.Errorf("failed to update user %s", err)
		}
		if rUser.Name != user.Name {
			return fmt.Errorf("failed to update user name before %s after %s", user.Name, rUser.Name)
		}
		if rUser.Metadata["Update"] != user.Metadata["Update"] {
			return fmt.Errorf("failed to update user metadata before %s after %s", user.Metadata["Update"], rUser.Metadata["Update"])
		}
		user = rUser
		user.Credentials.Identity = namesgenerator.Generate()
		rUser, err = s.UpdateUserIdentity(user, token)
		if err != nil {
			return fmt.Errorf("failed to update user idenitity %s", err)
		}
		if rUser.Credentials.Identity != user.Credentials.Identity {
			return fmt.Errorf("failed to update user identity before %s after %s", user.Credentials.Identity, rUser.Credentials.Identity)
		}
		user = rUser
		user.Tags = []string{namesgenerator.Generate()}
		rUser, err = s.UpdateUserTags(user, token)
		if err != nil {
			return fmt.Errorf("failed to update user tags %s", err)
		}
		if rUser.Tags[0] != user.Tags[0] {
			return fmt.Errorf("failed to update user tags before %s after %s", user.Tags[0], rUser.Tags[0])
		}
		user = rUser
		rUser, err = s.DisableUser(user.ID, token)
		if err != nil {
			return fmt.Errorf("failed to disable user %s", err)
		}
		if rUser.Status != "disabled" {
			return fmt.Errorf("failed to disable user before %s after %s", user.Status, rUser.Status)
		}
		user = rUser
		rUser, err = s.EnableUser(user.ID, token)
		if err != nil {
			return fmt.Errorf("failed to enable user %s", err)
		}
		if rUser.Status != "enabled" {
			return fmt.Errorf("failed to enable user before %s after %s", user.Status, rUser.Status)
		}
	}
	for _, group := range groups {
		group.Name = namesgenerator.Generate()
		group.Metadata = sdk.Metadata{"Update": namesgenerator.Generate()}
		rGroup, err := s.UpdateGroup(group, token)
		if err != nil {
			return fmt.Errorf("failed to update group %s", err)
		}
		if rGroup.Name != group.Name {
			return fmt.Errorf("failed to update group name before %s after %s", group.Name, rGroup.Name)
		}
		if rGroup.Metadata["Update"] != group.Metadata["Update"] {
			return fmt.Errorf("failed to update group metadata before %s after %s", group.Metadata["Update"], rGroup.Metadata["Update"])
		}
		group = rGroup
		rGroup, err = s.DisableGroup(group.ID, token)
		if err != nil {
			return fmt.Errorf("failed to disable group %s", err)
		}
		if rGroup.Status != "disabled" {
			return fmt.Errorf("failed to disable group before %s after %s", group.Status, rGroup.Status)
		}
		group = rGroup
		rGroup, err = s.EnableGroup(group.ID, token)
		if err != nil {
			return fmt.Errorf("failed to enable group %s", err)
		}
		if rGroup.Status != "enabled" {
			return fmt.Errorf("failed to enable group before %s after %s", group.Status, rGroup.Status)
		}
	}
	for _, thing := range things {
		thing.Name = namesgenerator.Generate()
		thing.Metadata = sdk.Metadata{"Update": namesgenerator.Generate()}
		rThing, err := s.UpdateThing(thing, token)
		if err != nil {
			return fmt.Errorf("failed to update thing %s", err)
		}
		if rThing.Name != thing.Name {
			return fmt.Errorf("failed to update thing name before %s after %s", thing.Name, rThing.Name)
		}
		if rThing.Metadata["Update"] != thing.Metadata["Update"] {
			return fmt.Errorf("failed to update thing metadata before %s after %s", thing.Metadata["Update"], rThing.Metadata["Update"])
		}
		thing = rThing
		rThing, err = s.UpdateThingSecret(thing.ID, thing.Credentials.Secret, token)
		if err != nil {
			return fmt.Errorf("failed to update thing secret %s", err)
		}
		if rThing.Credentials.Secret != thing.Credentials.Secret {
			return fmt.Errorf("failed to update thing secret before %s after %s", thing.Credentials.Secret, rThing.Credentials.Secret)
		}
		thing = rThing
		thing.Tags = []string{namesgenerator.Generate()}
		rThing, err = s.UpdateThingTags(thing, token)
		if err != nil {
			return fmt.Errorf("failed to update thing tags %s", err)
		}
		if rThing.Tags[0] != thing.Tags[0] {
			return fmt.Errorf("failed to update thing tags before %s after %s", thing.Tags[0], rThing.Tags[0])
		}
		thing = rThing
		rThing, err = s.DisableThing(thing.ID, token)
		if err != nil {
			return fmt.Errorf("failed to disable thing %s", err)
		}
		if rThing.Status != "disabled" {
			return fmt.Errorf("failed to disable thing before %s after %s", thing.Status, rThing.Status)
		}
		thing = rThing
		rThing, err = s.EnableThing(thing.ID, token)
		if err != nil {
			return fmt.Errorf("failed to enable thing %s", err)
		}
		if rThing.Status != "enabled" {
			return fmt.Errorf("failed to enable thing before %s after %s", thing.Status, rThing.Status)
		}
	}
	for _, channel := range channels {
		channel.Name = namesgenerator.Generate()
		channel.Metadata = sdk.Metadata{"Update": namesgenerator.Generate()}
		rChannel, err := s.UpdateChannel(channel, token)
		if err != nil {
			return fmt.Errorf("failed to update channel %s", err)
		}
		if rChannel.Name != channel.Name {
			return fmt.Errorf("failed to update channel name before %s after %s", channel.Name, rChannel.Name)
		}
		if rChannel.Metadata["Update"] != channel.Metadata["Update"] {
			return fmt.Errorf("failed to update channel metadata before %s after %s", channel.Metadata["Update"], rChannel.Metadata["Update"])
		}
		channel = rChannel
		rChannel, err = s.DisableChannel(channel.ID, token)
		if err != nil {
			return fmt.Errorf("failed to disable channel %s", err)
		}
		if rChannel.Status != "disabled" {
			return fmt.Errorf("failed to disable channel before %s after %s", channel.Status, rChannel.Status)
		}
		channel = rChannel
		rChannel, err = s.EnableChannel(channel.ID, token)
		if err != nil {
			return fmt.Errorf("failed to enable channel %s", err)
		}
		if rChannel.Status != "enabled" {
			return fmt.Errorf("failed to enable channel before %s after %s", channel.Status, rChannel.Status)
		}
	}
	return nil
}

func messaging(s sdk.SDK, conf Config, token string, things []sdk.Thing, channels []sdk.Channel) error {
	for _, thing := range things {
		for _, channel := range channels {
			if err := s.ConnectThing(thing.ID, channel.ID, token); err != nil {
				return fmt.Errorf("failed to connect thing %s to channel %s", thing.ID, channel.ID)
			}
		}
	}
	g := new(errgroup.Group)

	for i := uint64(0); i < conf.NumOfMsg; i++ {
		for _, thing := range things {
			thing := thing
			for _, channel := range channels {
				channel := channel
				g.Go(func() error {
					return sendHTTPMessage(s, thing, channel.ID)
				})
				g.Go(func() error {
					return sendCOAPMessage(thing, channel.ID)
				})
				g.Go(func() error {
					return sendMQTTMessage(thing, channel.ID)
				})
				g.Go(func() error {
					return sendWSMessage(conf, thing, channel.ID)
				})
			}
		}
	}

	return g.Wait()
}

func sendHTTPMessage(s sdk.SDK, thing sdk.Thing, chanID string) error {
	if err := s.SendMessage(chanID, msg, thing.Credentials.Secret); err != nil {
		return fmt.Errorf("http failed to send message from thing %s to channel %s: %v", thing.ID, chanID, err)
	}
	return nil
}

func sendCOAPMessage(thing sdk.Thing, chanID string) error {
	cmd := exec.Command("coap-cli", "post", fmt.Sprintf("channels/%s/messages", chanID), "-auth", thing.Credentials.Secret, "-d", msg)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("coap failed to send message from thing %s to channel %s: %v", thing.ID, chanID, err)
	}
	return nil
}

func sendMQTTMessage(thing sdk.Thing, chanID string) error {
	cmd := exec.Command("mosquitto_pub", "--id-prefix", "mainflux", "-u", thing.ID, "-P", thing.Credentials.Secret, "-t", fmt.Sprintf("channels/%s/messages", chanID), "-h", "localhost", "-m", msg)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mqtt failed to send message from thing %s to channel %s: %v", thing.ID, chanID, err)
	}
	return nil
}

func sendWSMessage(conf Config, thing sdk.Thing, chanID string) error {
	socketUrl := fmt.Sprintf("ws://%s:%s/channels/%s/messages", conf.Host, defWSPort, chanID)
	header := http.Header{"authorization": []string{thing.Credentials.Secret}}
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, header)
	if err != nil {
		return fmt.Errorf("unable to connect to websocket: %v", err)
	}
	defer conn.Close()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		return fmt.Errorf("mqtt failed to send message from thing %s to channel %s: %v", thing.ID, chanID, err)
	}
	return nil
}

func getIDS(objects interface{}) []string {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		panic("objects argument must be a slice")
	}
	ids := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		id := v.Index(i).FieldByName("ID").String()
		ids[i] = id
	}
	return ids
}
