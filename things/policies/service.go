package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

// Possible token types are access and refresh tokens.
const (
	ReadAction       = "m_read"
	WriteAction      = "m_write"
	ClientEntityType = "client"
	GroupEntityType  = "group"
)

type service struct {
	auth        upolicies.AuthServiceClient
	policies    Repository
	policyCache Cache
	idProvider  mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth upolicies.AuthServiceClient, p Repository, tcache clients.ClientCache, ccache Cache, idp mainflux.IDProvider) Service {
	return service{
		auth:        auth,
		policies:    p,
		policyCache: ccache,
		idProvider:  idp,
	}
}

func (svc service) Authorize(ctx context.Context, ar AccessRequest, entity string) (string, error) {
	// fetch from cache first
	p := Policy{
		Subject: ar.Subject,
		Object:  ar.Object,
	}
	policy, err := svc.policyCache.Get(ctx, p)
	if err == nil {
		for _, action := range policy.Actions {
			if action == ar.Action {
				return policy.Subject, nil
			}
		}
		return "", errors.ErrAuthorization
	}
	if !errors.Contains(err, errors.ErrNotFound) {
		return "", err
	}
	// fetch from repo as a fallback if not found in cache
	policy, err = svc.policies.RetrieveOne(ctx, p.Subject, p.Object)
	if err != nil {
		return "", err
	}

	// Replace Subject since AccessRequest Subject is Thing Key,
	// and Policy subject is Thing ID.
	policy.Subject = ar.Subject

	for _, action := range policy.Actions {
		if action == ar.Action {
			if err := svc.policyCache.Put(ctx, policy); err != nil {
				return policy.Subject, err
			}

			return policy.Subject, nil
		}
	}
	return "", errors.ErrAuthorization

}

func (svc service) AddPolicy(ctx context.Context, token string, p Policy) (Policy, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return Policy{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := p.Validate(); err != nil {
		return Policy{}, err
	}

	p.OwnerID = res.GetId()
	p.CreatedAt = time.Now()

	p, err = svc.policies.Save(ctx, p)
	if err != nil {
		return Policy{}, err
	}

	if err := svc.policyCache.Put(ctx, p); err != nil {
		return p, err
	}
	return p, nil
}

func (svc service) UpdatePolicy(ctx context.Context, token string, p Policy) (Policy, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return Policy{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := p.Validate(); err != nil {
		return Policy{}, err
	}
	if err := svc.checkAction(ctx, res.GetId(), p); err != nil {
		return Policy{}, err
	}
	p.UpdatedAt = time.Now()
	p.UpdatedBy = res.GetId()

	return svc.policies.Update(ctx, p)
}

func (svc service) ListPolicies(ctx context.Context, token string, pm Page) (PolicyPage, error) {
	if _, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token}); err != nil {
		return PolicyPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := pm.Validate(); err != nil {
		return PolicyPage{}, err
	}

	page, err := svc.policies.Retrieve(ctx, pm)
	if err != nil {
		return PolicyPage{}, err
	}

	return page, err
}

func (svc service) DeletePolicy(ctx context.Context, token string, p Policy) error {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := svc.checkAction(ctx, res.GetId(), p); err != nil {
		return err
	}
	if err := svc.policyCache.Remove(ctx, p); err != nil {
		return err
	}
	return svc.policies.Delete(ctx, p)
}

// checkAction check if client updating the policy has the sufficient priviledges.
// If the client is the owner of the policy.
// If the client is the admin.
func (svc service) checkAction(ctx context.Context, clientID string, p Policy) error {
	pm := Page{Subject: p.Subject, Object: p.Object, OwnerID: clientID, Total: 1, Offset: 0}
	page, err := svc.policies.Retrieve(ctx, pm)
	if err != nil {
		return err
	}
	if len(page.Policies) != 1 {
		return errors.ErrAuthorization
	}
	// If the client is the owner of the policy
	if page.Policies[0].OwnerID == clientID {
		return nil
	}

	// If the client is the admin
	req := &upolicies.AuthorizeReq{Sub: clientID, Obj: p.Object, Act: p.Actions[0], EntityType: "client"}
	if _, err := svc.auth.Authorize(ctx, req); err == nil {
		return nil
	}

	return errors.ErrAuthorization

}
