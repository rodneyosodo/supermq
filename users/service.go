package users

import (
	"context"
	"regexp"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/auth/policies"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	memberRelationKey = "c_add"
	authoritiesObjKey = "*"
	MyKey             = "mine"
)

var (
	// ErrMissingResetToken indicates malformed or missing reset token
	// for reseting password.
	ErrMissingResetToken = errors.New("missing reset token")

	// ErrRecoveryToken indicates error in generating password recovery token.
	ErrRecoveryToken = errors.New("failed to generate password recovery token")

	// ErrGetToken indicates error in getting signed token.
	ErrGetToken = errors.New("failed to fetch signed token")

	// ErrPasswordFormat indicates weak password.
	ErrPasswordFormat = errors.New("password does not meet the requirements")

	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid user status")

	// ErrEnableClient indicates error in enabling user.
	ErrEnableClient = errors.New("failed to enable user")

	// ErrDisableClient indicates error in disabling user.
	ErrDisableClient = errors.New("failed to disable user")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")
)

// // Service unites Clients and Group services.
// type Service interface {
// 	UserService
// 	jwt.TokenService
// }

type service struct {
	users      Repository
	hasher     Hasher
	email      Emailer
	auth       mainflux.AuthServiceClient
	idProvider mainflux.IDProvider
	passRegex  *regexp.Regexp
}

// New instantiates the users service implementation
func New(users Repository, hasher Hasher, auth mainflux.AuthServiceClient, e Emailer, idp mainflux.IDProvider, passRegex *regexp.Regexp) Service {
	return &service{
		users:      users,
		hasher:     hasher,
		auth:       auth,
		email:      e,
		idProvider: idp,
		passRegex:  passRegex,
	}
}
func (svc service) Register(ctx context.Context, token string, user User) (User, error) {
	if err := svc.checkAuthz(ctx, token); err != nil {
		return User{}, err
	}
	if err := user.Validate(); err != nil {
		return User{}, err
	}
	if !svc.passRegex.MatchString(user.Credentials.Secret) {
		return User{}, ErrPasswordFormat
	}

	uid, err := svc.idProvider.ID()
	if err != nil {
		return User{}, err
	}

	ir, err := svc.identify(ctx, token)
	if err != nil {
		return User{}, err
	}

	if ir.id != "" && user.Owner == "" {
		user.Owner = ir.id
	}

	hash, err := svc.hasher.Hash(user.Credentials.Secret)
	if err != nil {
		return User{}, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	if user.Status != DisabledStatus && user.Status != EnabledStatus {
		return User{}, apiutil.ErrInvalidStatus
	}

	user.ID = uid
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	user.Credentials.Secret = hash

	return svc.users.Save(ctx, user)
}

func (svc service) Login(ctx context.Context, user User) (string, error) {
	dbUser, err := svc.users.RetrieveByIdentity(ctx, user.Credentials.Identity)
	if err != nil {
		return "", errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := svc.hasher.Compare(user.Credentials.Secret, dbUser.Credentials.Secret); err != nil {
		return "", errors.Wrap(errors.ErrAuthentication, err)
	}
	return svc.issue(ctx, dbUser.ID, dbUser.Credentials.Identity, keys.LoginKey)
}

func (svc service) ViewUser(ctx context.Context, token string, id string) (User, error) {
	ir, err := svc.identify(ctx, token)
	if err != nil {
		return User{}, err
	}
	if ir.id == id {
		return svc.users.RetrieveByID(ctx, id)
	}

	if err := svc.authorize(ctx, "client", policies.Policy{Subject: ir.id, Object: id, Actions: []string{"c_list"}}); err != nil {
		return User{}, err
	}

	return svc.users.RetrieveByID(ctx, id)
}

func (svc service) ViewProfile(ctx context.Context, token string) (User, error) {
	ir, err := svc.identify(ctx, token)
	if err != nil {
		return User{}, err
	}

	return svc.users.RetrieveByIdentity(ctx, ir.email)
}

func (svc service) ListUsers(ctx context.Context, token string, pm Page) (UsersPage, error) {
	ir, err := svc.identify(ctx, token)
	if err != nil {
		return UsersPage{}, err
	}
	if pm.SharedBy == MyKey {
		pm.SharedBy = ir.id
	}
	if pm.OwnerID == MyKey {
		pm.OwnerID = ir.id
	}
	pm.Action = "c_list"
	clients, err := svc.users.RetrieveAll(ctx, pm)
	if err != nil {
		return UsersPage{}, err
	}
	for i, client := range clients.Users {
		if client.ID == ir.id {
			clients.Users = append(clients.Users[:i], clients.Users[i+1:]...)
			if clients.Total != 0 {
				clients.Total = clients.Total - 1
			}
		}
	}

	return clients, nil
}

func (svc service) UpdateUser(ctx context.Context, token string, user User) (User, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: user.ID, Actions: []string{"c_update"}}); err != nil {
		return User{}, err
	}

	user = User{
		ID:        user.ID,
		Name:      user.Name,
		Metadata:  user.Metadata,
		UpdatedAt: time.Now(),
	}

	return svc.users.Update(ctx, user)
}

func (svc service) UpdateUserTags(ctx context.Context, token string, user User) (User, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: user.ID, Actions: []string{"c_update"}}); err != nil {
		return User{}, err
	}

	user = User{
		ID:        user.ID,
		Tags:      user.Tags,
		UpdatedAt: time.Now(),
	}

	return svc.users.UpdateTags(ctx, user)
}

func (svc service) UpdateUserIdentity(ctx context.Context, token, id, identity string) (User, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: id, Actions: []string{"c_update"}}); err != nil {
		return User{}, err
	}

	user := User{
		ID: id,
		Credentials: Credentials{
			Identity: identity,
		},
	}
	return svc.users.UpdateIdentity(ctx, user)
}

func (svc service) GenerateResetToken(ctx context.Context, email, host string) error {
	user, err := svc.users.RetrieveByIdentity(ctx, email)
	if err != nil || user.Credentials.Identity == "" {
		return errors.ErrNotFound
	}
	t, err := svc.issue(ctx, user.ID, user.Credentials.Identity, keys.RecoveryKey)
	if err != nil {
		return errors.Wrap(ErrRecoveryToken, err)
	}
	return svc.SendPasswordReset(ctx, host, email, t)
}

func (svc service) ResetPassword(ctx context.Context, resetToken, password string) error {
	ir, err := svc.identify(ctx, resetToken)
	if err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}
	u, err := svc.users.RetrieveByIdentity(ctx, ir.email)
	if err != nil {
		return err
	}
	if u.Credentials.Identity == "" {
		return errors.ErrNotFound
	}
	if !svc.passRegex.MatchString(password) {
		return ErrPasswordFormat
	}
	password, err = svc.hasher.Hash(password)
	if err != nil {
		return err
	}
	user := User{
		Credentials: Credentials{
			Identity: ir.email,
			Secret:   password,
		},
	}
	if _, err := svc.users.UpdateSecret(ctx, user); err != nil {
		return err
	}
	return nil
}

func (svc service) ChangePassword(ctx context.Context, token, oldSecret, newSecret string) (User, error) {
	ir, err := svc.identify(ctx, token)
	if err != nil {
		return User{}, err
	}
	if !svc.passRegex.MatchString(newSecret) {
		return User{}, ErrPasswordFormat
	}

	dbClient, err := svc.users.RetrieveByID(ctx, ir.id)
	if err != nil {
		return User{}, err
	}
	if err := svc.hasher.Compare(oldSecret, dbClient.Credentials.Secret); err != nil {
		return User{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	user := User{
		Credentials: Credentials{
			Identity: ir.email,
			Secret:   oldSecret,
		},
	}
	if _, err := svc.Login(ctx, user); err != nil {
		return User{}, errors.ErrAuthentication
	}
	user, err = svc.users.RetrieveByIdentity(ctx, ir.email)
	if err != nil || user.Credentials.Identity == "" {
		return User{}, errors.ErrNotFound
	}
	newSecret, err = svc.hasher.Hash(newSecret)
	if err != nil {
		return User{}, err
	}
	dbClient.Credentials.Secret = newSecret
	return svc.users.UpdateSecret(ctx, dbClient)
}

func (svc service) SendPasswordReset(_ context.Context, host, email, token string) error {
	to := []string{email}
	return svc.email.SendPasswordReset(to, host, token)
}

func (svc service) UpdateUserOwner(ctx context.Context, token string, user User) (User, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: user.ID, Actions: []string{"c_update"}}); err != nil {
		return User{}, err
	}

	user = User{
		ID:        user.ID,
		Owner:     user.Owner,
		UpdatedAt: time.Now(),
	}

	return svc.users.UpdateOwner(ctx, user)
}

func (svc service) EnableUser(ctx context.Context, token, id string) (User, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: id, Actions: []string{"c_delete"}}); err != nil {
		return User{}, err
	}
	user, err := svc.changeUserStatus(ctx, id, EnabledStatus)
	if err != nil {
		return User{}, errors.Wrap(ErrEnableClient, err)
	}

	return user, nil
}

func (svc service) DisableUser(ctx context.Context, token, id string) (User, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: id, Actions: []string{"c_delete"}}); err != nil {
		return User{}, err
	}
	user, err := svc.changeUserStatus(ctx, id, DisabledStatus)
	if err != nil {
		return User{}, errors.Wrap(ErrDisableClient, err)
	}

	return user, nil
}

func (svc service) ListMembers(ctx context.Context, token, groupID string, pm Page) (MembersPage, error) {
	ir, err := svc.identify(ctx, token)
	if err != nil {
		return MembersPage{}, err
	}
	pm.Subject = ir.id
	pm.Action = "g_list"

	return svc.users.Members(ctx, groupID, pm)
}

func (svc service) changeUserStatus(ctx context.Context, id string, status Status) (User, error) {
	dbClient, err := svc.users.RetrieveByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	if dbClient.Status == status {
		return User{}, ErrStatusAlreadyAssigned
	}

	return svc.users.ChangeStatus(ctx, id, status)
}

func (svc service) authorize(ctx context.Context, entityType string, p policies.Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	ir, err := svc.identify(ctx, p.Subject)
	if err != nil {
		return err
	}
	p.Subject = ir.id
	return svc.authorize(ctx, entityType, p)
}

type userIdentity struct {
	id    string
	email string
}

func (svc service) identify(ctx context.Context, token string) (userIdentity, error) {
	identity, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return userIdentity{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	return userIdentity{identity.Id, identity.Email}, nil
}

func (svc service) checkAuthz(ctx context.Context, token string) error {
	if token == "" {
		return errors.ErrAuthentication
	}
	pr := policies.Policy{
		Subject: "user",
		Object:  "*",
		Actions: []string{"create"},
	}
	if err := svc.authorize(ctx, "client", pr); err == nil {
		return nil
	}
	ir, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	pr = policies.Policy{
		Subject: ir.id,
		Object:  authoritiesObjKey,
		Actions: []string{memberRelationKey},
	}
	return svc.authorize(ctx, "client", pr)
}

// Auth helpers
func (svc service) issue(ctx context.Context, id, email string, keyType uint32) (string, error) {
	key, err := svc.auth.Issue(ctx, &mainflux.IssueReq{Id: id, Email: email, Type: keyType})
	if err != nil {
		return "", errors.Wrap(errors.ErrNotFound, err)
	}
	return key.GetValue(), nil
}
