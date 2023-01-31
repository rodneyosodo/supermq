package groups

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/auth/policies"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

// Possible token types are access and refresh tokens.
const (
	RefreshToken = "refresh"
	AccessToken  = "access"
	MyKey        = "mine"
)

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid groups status")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")
)

// Service unites Clients and Group services.
type Service interface {
	GroupService
}

type service struct {
	keys       keys.KeyRepository
	groups     GroupRepository
	tokenizer  keys.Tokenizer
	policies   policies.PolicyRepository
	idProvider mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(keys keys.KeyRepository, g GroupRepository, p policies.PolicyRepository, tokenizer keys.Tokenizer, idp mainflux.IDProvider) Service {
	return service{
		keys:       keys,
		groups:     g,
		policies:   p,
		idProvider: idp,
		tokenizer:  tokenizer,
	}
}

func (svc service) CreateGroup(ctx context.Context, token string, g Group) (Group, error) {
	ownerID, err := svc.identify(ctx, token)
	if err != nil {
		return Group{}, err
	}
	groupID, err := svc.idProvider.ID()
	if err != nil {
		return Group{}, err
	}
	if g.Status != EnabledStatus && g.Status != DisabledStatus {
		return Group{}, apiutil.ErrInvalidStatus
	}
	if g.OwnerID == "" {
		g.OwnerID = ownerID
	}

	g.ID = groupID
	g.CreatedAt = time.Now()
	g.UpdatedAt = g.CreatedAt
	return svc.groups.Save(ctx, g)
}

func (svc service) ViewGroup(ctx context.Context, token string, id string) (Group, error) {
	if err := svc.authorize(ctx, "group", policies.Policy{Subject: token, Object: id, Actions: []string{"g_list"}}); err != nil {
		return Group{}, err
	}

	return svc.groups.RetrieveByID(ctx, id)
}

func (svc service) ListGroups(ctx context.Context, token string, gm GroupsPage) (GroupsPage, error) {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return GroupsPage{}, err
	}
	gm.Subject = id
	gm.OwnerID = id
	gm.Action = "g_list"
	return svc.groups.RetrieveAll(ctx, gm)
}

func (svc service) ListMemberships(ctx context.Context, token, clientID string, gm GroupsPage) (MembershipsPage, error) {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return MembershipsPage{}, err
	}
	gm.Subject = id
	gm.Action = "g_list"
	return svc.groups.Memberships(ctx, clientID, gm)
}

func (svc service) UpdateGroup(ctx context.Context, token string, g Group) (Group, error) {
	if err := svc.authorize(ctx, "group", policies.Policy{Subject: token, Object: g.ID, Actions: []string{"g_update"}}); err != nil {
		return Group{}, err
	}
	g.UpdatedAt = time.Now()

	return svc.groups.Update(ctx, g)
}

func (svc service) EnableGroup(ctx context.Context, token, id string) (Group, error) {
	if err := svc.authorize(ctx, "group", policies.Policy{Subject: token, Object: id, Actions: []string{"g_delete"}}); err != nil {
		return Group{}, err
	}
	group, err := svc.changeGroupStatus(ctx, id, EnabledStatus)
	if err != nil {
		return Group{}, err
	}
	return group, nil
}

func (svc service) DisableGroup(ctx context.Context, token, id string) (Group, error) {
	if err := svc.authorize(ctx, "group", policies.Policy{Subject: token, Object: id, Actions: []string{"g_delete"}}); err != nil {
		return Group{}, err
	}
	group, err := svc.changeGroupStatus(ctx, id, DisabledStatus)
	if err != nil {
		return Group{}, err
	}
	return group, nil
}

func (svc service) authorize(ctx context.Context, entityType string, p policies.Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	id, err := svc.identify(ctx, p.Subject)
	if err != nil {
		return err
	}
	p.Subject = id
	return svc.policies.Evaluate(ctx, entityType, p)
}

func (svc service) changeGroupStatus(ctx context.Context, id string, status Status) (Group, error) {
	dbGroup, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return Group{}, err
	}
	if dbGroup.Status == status {
		return Group{}, ErrStatusAlreadyAssigned
	}

	return svc.groups.ChangeStatus(ctx, id, status)
}

func (svc service) identify(ctx context.Context, token string) (string, error) {
	key, err := svc.tokenizer.Parse(token)
	if err == keys.ErrAPIKeyExpired {
		err = svc.keys.Remove(ctx, key.IssuerID, key.ID)
		return "", errors.Wrap(keys.ErrAPIKeyExpired, err)
	}
	if err != nil {
		return "", err
	}

	switch key.Type {
	case keys.RecoveryKey, keys.LoginKey:
		return key.IssuerID, nil
	case keys.APIKey:
		_, err := svc.keys.Retrieve(context.TODO(), key.IssuerID, key.ID)
		if err != nil {
			return "", errors.ErrAuthentication
		}
		return key.IssuerID, nil
	default:
		return "", errors.ErrAuthentication
	}
}
