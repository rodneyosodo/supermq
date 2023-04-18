// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package policies

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gookit/color"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

const (
	defPass = "12345678"
)

var (
	seed = time.Now().UTC().UnixNano()
)

// Config - test configuration.
type Config struct {
	Host  string
	SSL   bool
	CA    string
	CAKey string
}

func init() {
	rand.Seed(seed)
}

func init() {
	// cmd := exec.Command("docker", "exec", "-it", "mainflux-users-db", "psql", "-U", "mainflux", "-d", "users", "-c", `"delete from policies; delete from groups; delete from clients;"`)
	// if err := cmd.Run(); err != nil {
	// 	log.Fatal("failed to remove esisting things in the database with error: ", err)
	// }
}

// Test - function that does actual end to end testing.
// | subject | object | actions                                     |
// | ------- | ------ | ------------------------------------------- |
// | clientA | groupA | ["g_add", "g_list", "g_update", "g_delete"] |
// | clientB | groupA | ["c_list", "c_update", "c_delete"]          |
// | clientC | groupA | ["c_update"]                                |
// | clientD | groupA | ["c_list"]                                  |
// | clientE | groupB | ["c_list", "c_update", "c_delete"]          |
// | clientF | groupB | ["c_update"]                                |
// | clientD | groupB | ["c_list"]                                  |
func Test(conf Config) {
	sdkConf := sdk.Config{
		ThingsURL:       fmt.Sprintf("http://%s", conf.Host),
		UsersURL:        fmt.Sprintf("http://%s", conf.Host),
		MsgContentType:  sdk.CTJSONSenML,
		TLSVerification: false,
	}

	s := sdk.NewSDK(sdkConf)

	magenta := color.FgLightMagenta.Render

	token, err := createUser(s)
	if err != nil {
		errExit(err)
	}
	color.Success.Printf("created user with token %s\n", magenta(token))

	users, err := createUsers(s, token)
	if err != nil {
		errExit(err)
	}
	color.Success.Printf("created users of ids:\n%s\n", magenta(getIDS(users)))

	groups, err := createGroups(s, token)
	if err != nil {
		errExit(err)
	}
	color.Success.Printf("created groups of ids:\n%s\n", magenta(getIDS(groups)))

	if err := createPolicies(s, token, users, groups); err != nil {
		errExit(err)
	}
	color.Success.Println("created policies for users, groups")

	if err := validateClientAPolicies(s, users, groups); err != nil {
		errExit(fmt.Errorf("failed to validate client a policies: %w", err))
	}
	color.Success.Println("validated policies for client a")

	if err := validateClientBPolicies(s, users, groups); err != nil {
		errExit(fmt.Errorf("failed to validate client b policies: %w", err))
	}
	color.Success.Println("validated policies for client b")

	if err := validateClientCPolicies(s, users, groups); err != nil {
		errExit(fmt.Errorf("failed to validate client c policies: %w", err))
	}
	color.Success.Println("validated policies for client c")

	if err := validateClientDPolicies(s, users, groups); err != nil {
		errExit(fmt.Errorf("failed to validate client d policies: %w", err))
	}
	color.Success.Println("validated policies for client d")

	if err := validateClientEPolicies(s, users, groups); err != nil {
		errExit(fmt.Errorf("failed to validate client e policies: %w", err))
	}
	color.Success.Println("validated policies for client e")

	if err := validateClientFPolicies(s, users, groups); err != nil {
		errExit(fmt.Errorf("failed to validate client f policies: %w", err))
	}
	color.Success.Println("validated policies for client f")
}

func errExit(err error) {
	color.Error.Println(err.Error())
	os.Exit(1)
}

func createUser(s sdk.SDK) (string, error) {
	user := sdk.User{
		Name: "policies-admin",
		Credentials: sdk.Credentials{
			Identity: "admin@policies.com",
			Secret:   defPass,
		},
		Status: "enabled",
	}

	pass := user.Credentials.Secret

	user, err := s.CreateUser(user, "")
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	user.Credentials.Secret = pass
	token, err := s.CreateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to login user: %w", err)
	}

	return token.AccessToken, nil
}

func createUsers(s sdk.SDK, token string) ([]sdk.User, error) {
	var err error
	users := []sdk.User{}

	for i := 'a'; i <= 'f'; i++ {
		user := sdk.User{
			Name: fmt.Sprintf("client%c", i),
			Credentials: sdk.Credentials{
				Identity: fmt.Sprintf("client%c@policies.com", i),
				Secret:   defPass},
			Status: "enabled",
		}
		user, err = s.CreateUser(user, token)
		if err != nil {
			return []sdk.User{}, fmt.Errorf("failed to create the users: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func createGroups(s sdk.SDK, token string) ([]sdk.Group, error) {
	var err error
	groups := []sdk.Group{}

	parentID := ""
	for i := 'a'; i <= 'b'; i++ {
		group := sdk.Group{
			Name:     fmt.Sprintf("group%c", i),
			ParentID: parentID,
			Status:   "enabled",
		}

		group, err = s.CreateGroup(group, token)
		if err != nil {
			return []sdk.Group{}, fmt.Errorf("failed to create the group: %w", err)
		}
		groups = append(groups, group)
		parentID = group.ID
	}

	return groups, nil
}

func createPolicies(s sdk.SDK, token string, users []sdk.User, groups []sdk.Group) error {
	// create policies for group A
	aa := sdk.Policy{
		Subject: users[0].ID,
		Object:  groups[0].ID,
		Actions: []string{"g_add", "g_list", "g_update", "g_delete"},
	}
	if err := s.AddPolicy(aa, token); err != nil {
		return err
	}
	ba := sdk.Policy{
		Subject: users[1].ID,
		Object:  groups[0].ID,
		Actions: []string{"c_list", "c_update", "c_delete"},
	}
	if err := s.AddPolicy(ba, token); err != nil {
		return err
	}
	ca := sdk.Policy{
		Subject: users[2].ID,
		Object:  groups[0].ID,
		Actions: []string{"c_update"},
	}
	if err := s.AddPolicy(ca, token); err != nil {
		return err
	}
	da := sdk.Policy{
		Subject: users[3].ID,
		Object:  groups[0].ID,
		Actions: []string{"c_list"},
	}
	if err := s.AddPolicy(da, token); err != nil {
		return err
	}

	// create policies for group B
	eb := sdk.Policy{
		Subject: users[4].ID,
		Object:  groups[1].ID,
		Actions: []string{"c_list", "c_update", "c_delete"},
	}
	if err := s.AddPolicy(eb, token); err != nil {
		return err
	}
	fb := sdk.Policy{
		Subject: users[5].ID,
		Object:  groups[1].ID,
		Actions: []string{"c_update"},
	}
	if err := s.AddPolicy(fb, token); err != nil {
		return err
	}
	db := sdk.Policy{
		Subject: users[3].ID,
		Object:  groups[1].ID,
		Actions: []string{"c_list"},
	}
	if err := s.AddPolicy(db, token); err != nil {
		return err
	}
	return nil
}

func validateClientAPolicies(s sdk.SDK, users []sdk.User, groups []sdk.Group) error {
	// when `clientA` lists groups `groupA` will be listed
	// clientA can add members to `groupA`
	// clientA can update `groupA`
	// clientA can change the status of `groupA`

	users[0].Credentials.Secret = defPass
	token, err := s.CreateToken(users[0])
	if err != nil {
		return err
	}

	groupsA, err := s.Groups(sdk.PageMetadata{Offset: uint64(0), Limit: uint64(10)}, token.AccessToken)
	if err != nil {
		return err
	}
	if len(groupsA.Groups) != 1 {
		return fmt.Errorf("expected 1 group got %d", len(groupsA.Groups))
	}
	if groupsA.Total != 1 {
		return fmt.Errorf("expected 1 group got %d", groupsA.Total)
	}
	if groupsA.Groups[0].ID != groups[0].ID {
		return fmt.Errorf("expected groupa got %s", groupsA.Groups[0].ID)
	}
	groupA, err := s.UpdateGroup(sdk.Group{ID: groups[0].ID, Name: "updatedByA", Description: "updatedByA"}, token.AccessToken)
	if err != nil {
		return err
	}
	if groupA.Name != "updatedByA" {
		return fmt.Errorf("expected updatedByA got %s", groupA.Name)
	}
	if groupA.Description != "updatedByA" {
		return fmt.Errorf("expected updatedByA got %s", groupA.Description)
	}
	groupA, err = s.DisableGroup(groups[0].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if groupA.Status != sdk.DisabledStatus {
		return fmt.Errorf("expected disabled got %s", groupA.Status)
	}
	groupA, err = s.EnableGroup(groups[0].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if groupA.Status != sdk.EnabledStatus {
		return fmt.Errorf("expected enabled got %s", groupA.Status)
	}
	return nil
}

func validateClientBPolicies(s sdk.SDK, users []sdk.User, _ []sdk.Group) error {
	// when client B list clients they will list `clientA`, `clientC` and `clientD`
	// since they are connected in the same group `groupA` and they have `c_list` actions
	// clientB can update clients connected to the same group they are connected in
	// i.e they can update `clientA`, `clientC` and `clientD` since they are in the same `groupA`
	// clientB can change clients status of clients connected to the same group they are connected in
	// i.e they are able to change the status of `clientA`, `clientC` and `clientD` since they are in the same group `groupA`

	users[1].Credentials.Secret = defPass
	token, err := s.CreateToken(users[1])
	if err != nil {
		return err
	}

	usersA, err := s.Users(sdk.PageMetadata{Offset: uint64(0), Limit: uint64(10), Visibility: "shared"}, token.AccessToken)
	if err != nil {
		return err
	}
	if len(usersA.Users) != 3 {
		return fmt.Errorf("expected 3 users got %d", len(usersA.Users))
	}
	if usersA.Total != 3 {
		return fmt.Errorf("expected 3 users got %d", usersA.Total)
	}
	if usersA.Users[0].Credentials.Identity != "clienta@policies.com" {
		return fmt.Errorf("expected clienta@policies.com  got %s", usersA.Users[0].Credentials.Identity)
	}
	if usersA.Users[1].Credentials.Identity != "clientc@policies.com" {
		return fmt.Errorf("expected clientc@policies.com  got %s", usersA.Users[1].Credentials.Identity)
	}
	if usersA.Users[2].Credentials.Identity != "clientd@policies.com" {
		return fmt.Errorf("expected clientad@policies.com  got %s", usersA.Users[2].Credentials.Identity)
	}
	userA, err := s.UpdateUser(sdk.User{ID: users[0].ID, Name: "updatedA"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userA.Name != "updatedA" {
		return fmt.Errorf("expected updatedA for userA, got %s", userA.Name)
	}
	userC, err := s.UpdateUser(sdk.User{ID: users[2].ID, Name: "updatedA"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userC.Name != "updatedA" {
		return fmt.Errorf("expected updatedA for userC, got %s", userC.Name)
	}
	userD, err := s.UpdateUser(sdk.User{ID: users[3].ID, Name: "updatedA"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userD.Name != "updatedA" {
		return fmt.Errorf("expected updatedA for userD, got %s", userD.Name)
	}

	userA, err = s.DisableUser(users[0].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if userA.Status != sdk.DisabledStatus {
		return fmt.Errorf("expected disabled for userA, got %s", userA.Status)
	}
	userA, err = s.EnableUser(users[0].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if userA.Status != sdk.EnabledStatus {
		return fmt.Errorf("expected enabled for userA, got %s", userA.Status)
	}
	userC, err = s.DisableUser(users[2].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if userC.Status != sdk.DisabledStatus {
		return fmt.Errorf("expected disabled for userC, got %s", userC.Status)
	}
	userC, err = s.EnableUser(users[2].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if userC.Status != sdk.EnabledStatus {
		return fmt.Errorf("expected enabled for userC, got %s", userC.Status)
	}
	userD, err = s.DisableUser(users[3].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if userD.Status != sdk.DisabledStatus {
		return fmt.Errorf("expected disabled for userD, got %s", userD.Status)
	}
	userD, err = s.EnableUser(users[3].ID, token.AccessToken)
	if err != nil {
		return err
	}
	if userD.Status != sdk.EnabledStatus {
		return fmt.Errorf("expected enabled for userD, got %s", userD.Status)
	}
	return nil
}

func validateClientCPolicies(s sdk.SDK, users []sdk.User, _ []sdk.Group) error {
	// client C can update clients connected to the same group they are connected in
	// i.e they can update `clientA`, `clientB` and `clientD` since they are in the same `groupA`

	users[2].Credentials.Secret = defPass
	token, err := s.CreateToken(users[2])
	if err != nil {
		return err
	}

	userA, err := s.UpdateUser(sdk.User{ID: users[0].ID, Name: "updatedByC"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userA.Name != "updatedByC" {
		return fmt.Errorf("expected updatedByC got %s", userA.Name)
	}
	userB, err := s.UpdateUser(sdk.User{ID: users[1].ID, Name: "updatedByC"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userB.Name != "updatedByC" {
		return fmt.Errorf("expected updatedByC got %s", userB.Name)
	}
	userD, err := s.UpdateUser(sdk.User{ID: users[3].ID, Name: "updatedByC"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userD.Name != "updatedByC" {
		return fmt.Errorf("expected updatedByC got %s", userD.Name)
	}

	return nil
}

func validateClientDPolicies(s sdk.SDK, users []sdk.User, _ []sdk.Group) error {
	// when client D list clients they will list `clientA`, `clientB` and `clientC`
	// since they are connected in the same group `groupA` and they have `c_list` actions
	// and also `clientE` and `clientF` since they are connected to the same group `groupB`
	// and they have `c_list` actions

	users[3].Credentials.Secret = defPass
	token, err := s.CreateToken(users[3])
	if err != nil {
		return err
	}

	usersA, err := s.Users(sdk.PageMetadata{Offset: uint64(0), Limit: uint64(10), Visibility: "shared"}, token.AccessToken)
	if err != nil {
		return err
	}
	if len(usersA.Users) != 6 {
		return fmt.Errorf("expected 6 users got %d", len(usersA.Users))
	}
	if usersA.Total != 6 {
		return fmt.Errorf("expected 6 users got %d", usersA.Total)
	}
	if usersA.Users[0].Credentials.Identity != "clienta@policies.com" {
		return fmt.Errorf("expected clienta@policies.com  got %s", usersA.Users[0].Credentials.Identity)
	}
	if usersA.Users[1].Credentials.Identity != "clientb@policies.com" {
		return fmt.Errorf("expected clientb@policies.com  got %s", usersA.Users[1].Credentials.Identity)
	}
	if usersA.Users[2].Credentials.Identity != "clientc@policies.com" {
		return fmt.Errorf("expected clientc@policies.com  got %s", usersA.Users[2].Credentials.Identity)
	}
	if usersA.Users[4].Credentials.Identity != "cliente@policies.com" {
		return fmt.Errorf("expected cliente@policies.com  got %s", usersA.Users[4].Credentials.Identity)
	}
	if usersA.Users[5].Credentials.Identity != "clientf@policies.com" {
		return fmt.Errorf("expected clientf@policies.com  got %s", usersA.Users[5].Credentials.Identity)
	}
	return nil
}

func validateClientEPolicies(s sdk.SDK, users []sdk.User, _ []sdk.Group) error {
	// when clientE list clients they will list `clientF` and `clientD`
	// since they are connected in the same group `groupB` and they have `c_list` actions
	// client E can update clients connected to the same group they are connected in
	// i.e they can update `clientF` and `clientD` since they are in the same `groupB`
	users[4].Credentials.Secret = defPass
	token, err := s.CreateToken(users[4])
	if err != nil {
		return err
	}

	usersA, err := s.Users(sdk.PageMetadata{Offset: uint64(0), Limit: uint64(10), Visibility: "shared"}, token.AccessToken)
	if err != nil {
		return err
	}
	if len(usersA.Users) != 2 {
		return fmt.Errorf("expected 2 users got %d", len(usersA.Users))
	}
	if usersA.Total != 2 {
		return fmt.Errorf("expected 2 users got %d", usersA.Total)
	}
	if usersA.Users[0].Credentials.Identity != "clientd@policies.com" {
		return fmt.Errorf("expected clientd@policies.com  got %s", usersA.Users[1].Credentials.Identity)
	}
	if usersA.Users[1].Credentials.Identity != "clientf@policies.com" {
		return fmt.Errorf("expected clientf@policies.com  got %s", usersA.Users[0].Credentials.Identity)
	}

	userA, err := s.UpdateUser(sdk.User{ID: users[3].ID, Name: "updatedByE"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userA.Name != "updatedByE" {
		return fmt.Errorf("expected updatedByE got %s", userA.Name)
	}
	userC, err := s.UpdateUser(sdk.User{ID: users[5].ID, Name: "updatedByE"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userC.Name != "updatedByE" {
		return fmt.Errorf("expected updatedByE got %s", userC.Name)
	}

	return nil
}

func validateClientFPolicies(s sdk.SDK, users []sdk.User, _ []sdk.Group) error {
	// client F can update clients connected to the same group they are connected in
	// i.e they can update `clientE`, and `clientD` since they are in the same `groupB`

	users[5].Credentials.Secret = defPass
	token, err := s.CreateToken(users[5])
	if err != nil {
		return err
	}

	userA, err := s.UpdateUser(sdk.User{ID: users[3].ID, Name: "updatedByF"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userA.Name != "updatedByF" {
		return fmt.Errorf("expected updatedByF got %s", userA.Name)
	}
	userC, err := s.UpdateUser(sdk.User{ID: users[4].ID, Name: "updatedByF"}, token.AccessToken)
	if err != nil {
		return err
	}
	if userC.Name != "updatedByF" {
		return fmt.Errorf("expected updatedByF got %s", userC.Name)
	}

	return nil
}

// getIDS returns a list of IDs of the given objects.
func getIDS(objects interface{}) string {
	v := reflect.ValueOf(objects)
	if v.Kind() != reflect.Slice {
		panic("objects argument must be a slice")
	}
	ids := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		id := v.Index(i).FieldByName("ID").String()
		ids[i] = id
	}
	idList := strings.Join(ids, "\n")

	return idList
}
