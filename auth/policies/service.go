package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth/keys"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

// Possible token types are access and refresh tokens.
const (
	AccessToken = "access"
)

// Service unites Clients and Group services.
type Service interface {
	PolicyService
}

type service struct {
	policies   PolicyRepository
	idProvider mainflux.IDProvider
	tokens     keys.Tokenizer
	keys       keys.KeyRepository
}

// NewService returns a new Clients service implementation.
func NewService(p PolicyRepository, t keys.Tokenizer, k keys.KeyRepository, idp mainflux.IDProvider) Service {
	return service{
		policies:   p,
		tokens:     t,
		keys:       k,
		idProvider: idp,
	}
}

func (svc service) Authorize(ctx context.Context, entityType string, p Policy) error {
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
func (svc service) UpdatePolicy(ctx context.Context, token string, p Policy) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if err := svc.checkActionRank(ctx, id, p); err != nil {
		return err
	}
	p.UpdatedAt = time.Now()

	return svc.policies.Update(ctx, p)
}

func (svc service) AddPolicy(ctx context.Context, token string, p Policy) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := p.Validate(); err != nil {
		return err
	}

	page, err := svc.policies.Retrieve(ctx, Page{Subject: p.Subject, Object: p.Object})
	if err != nil {
		return err
	}
	if len(page.Policies) != 0 {
		return svc.policies.Update(ctx, p)
	}
	if err := svc.checkActionRank(ctx, id, p); err != nil {
		return err
	}
	p.OwnerID = id
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt

	return svc.policies.Save(ctx, p)
}

func (svc service) DeletePolicy(ctx context.Context, token string, p Policy) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := svc.checkActionRank(ctx, id, p); err != nil {
		return err
	}

	return svc.policies.Delete(ctx, p)
}

func (svc service) ListPolicy(ctx context.Context, token string, pm Page) (PolicyPage, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return PolicyPage{}, err
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

// checkActionRank check if an action is in the provide list of actions
func (svc service) checkActionRank(ctx context.Context, clientID string, p Policy) error {
	page, err := svc.policies.Retrieve(ctx, Page{Subject: clientID, Object: p.Object})
	if err != nil {
		return err
	}
	if len(page.Policies) != 0 {
		for _, a := range p.Actions {
			var found = false
			for _, v := range page.Policies[0].Actions {
				if v == a {
					found = true
					break
				}
			}
			if !found {
				return apiutil.ErrHigherPolicyRank
			}
		}
	}

	return nil

}

func (svc service) identify(ctx context.Context, token string) (string, error) {
	key, err := svc.tokens.Parse(token)
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
