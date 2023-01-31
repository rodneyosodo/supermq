package users

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"golang.org/x/net/idna"
)

const (
	maxLocalLen  = 64
	maxDomainLen = 255
	maxTLDLen    = 24 // longest TLD currently in existence

	atSeparator  = "@"
	dotSeparator = "."
)

var (
	userRegexp    = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp    = regexp.MustCompile(`^[^\s]+\.[^\s]+$`)
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

// Credentials represent client credentials: its
// "identity" which can be a username, email, generated name;
// and "secret" which can be a password or access token.
type Credentials struct {
	Identity string `json:"email"`    // username or generated login ID
	Secret   string `json:"password"` // password or token
}

// User represents generic Client.
type User struct {
	ID          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Owner       string      `json:"owner,omitempty"` // nullable
	Credentials Credentials `json:"credentials"`
	Metadata    Metadata    `json:"metadata,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Status      Status      `json:"status"` // 1 for enabled, 0 for disabled
}

// ClientsPage contains page related metadata as well as list
// of Clients that belong to the page.
type UsersPage struct {
	Page
	Users []User
}

// MembersPage contains page related metadata as well as list of members that
// belong to this page.
type MembersPage struct {
	Page
	Members []User
}

// Repository specifies an account persistence API.
type Repository interface {
	// Save persists the client account. A non-nil error is returned to indicate
	// operation failure.
	Save(ctx context.Context, client User) (User, error)

	// RetrieveByID retrieves client by its unique ID.
	RetrieveByID(ctx context.Context, id string) (User, error)

	// RetrieveByIdentity retrieves client by its unique credentials
	RetrieveByIdentity(ctx context.Context, identity string) (User, error)

	// RetrieveAll retrieves all clients.
	RetrieveAll(ctx context.Context, pm Page) (UsersPage, error)

	// Members retrieves everything that is assigned to a group identified by groupID.
	Members(ctx context.Context, groupID string, pm Page) (MembersPage, error)

	// Update updates the client name and metadata.
	Update(ctx context.Context, client User) (User, error)

	// UpdateTags updates the client tags.
	UpdateTags(ctx context.Context, client User) (User, error)

	// UpdateIdentity updates identity for client with given id.
	UpdateIdentity(ctx context.Context, client User) (User, error)

	// UpdateSecret updates secret for client with given identity.
	UpdateSecret(ctx context.Context, client User) (User, error)

	// UpdateOwner updates owner for client with given id.
	UpdateOwner(ctx context.Context, client User) (User, error)

	// ChangeStatus changes client status to enabled or disabled
	ChangeStatus(ctx context.Context, id string, status Status) (User, error)
}

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Register creates new client. In case of the failed registration, a
	// non-nil error value is returned.
	Register(ctx context.Context, token string, user User) (User, error)

	// Login authenticates the user given its credentials. Successful
	// authentication generates new access token. Failed invocations are
	// identified by the non-nil error values in the response.
	Login(ctx context.Context, user User) (string, error)

	// ViewUser retrieves client info for a given client ID and an authorized token.
	ViewUser(ctx context.Context, token, id string) (User, error)

	// ViewProfile retrieves user info for a given token.
	ViewProfile(ctx context.Context, token string) (User, error)

	// ListUsers retrieves clients list for a valid auth token.
	ListUsers(ctx context.Context, token string, pm Page) (UsersPage, error)

	// ListMembers retrieves everything that is assigned to a group identified by groupID.
	ListMembers(ctx context.Context, token, groupID string, pm Page) (MembersPage, error)

	// UpdateUser updates the client's name and metadata.
	UpdateUser(ctx context.Context, token string, client User) (User, error)

	// UpdateUserTags updates the client's tags.
	UpdateUserTags(ctx context.Context, token string, client User) (User, error)

	// UpdateUserIdentity updates the client's identity
	UpdateUserIdentity(ctx context.Context, token, id, identity string) (User, error)

	// UpdateUserOwner updates the client's owner.
	UpdateUserOwner(ctx context.Context, token string, client User) (User, error)

	// GenerateResetToken email where mail will be sent.
	// host is used for generating reset link.
	GenerateResetToken(ctx context.Context, email, host string) error

	// ChangePassword change users password for authenticated user.
	ChangePassword(ctx context.Context, authToken, password, oldPassword string) (User, error)

	// ResetPassword change users password in reset flow.
	// token can be authentication token or password reset token.
	ResetPassword(ctx context.Context, resetToken, password string) error

	// SendPasswordReset sends reset password link to email.
	SendPasswordReset(ctx context.Context, host, email, token string) error

	// EnableUser logically enableds the client identified with the provided ID
	EnableUser(ctx context.Context, token, id string) (User, error)

	// DisableUser logically disables the user identified with the provided ID
	DisableUser(ctx context.Context, token, id string) (User, error)
}

// Validate returns an error if user representation is invalid.
func (u User) Validate() error {
	if !isEmail(u.Credentials.Identity) {
		return errors.ErrMalformedEntity
	}
	return nil
}

// Custom Marshaller for Client
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Custom Unmarshaller for Client
func (s *Status) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ToStatus(str)
	*s = val
	return err
}

func isEmail(email string) bool {
	if email == "" {
		return false
	}

	es := strings.Split(email, atSeparator)
	if len(es) != 2 {
		return false
	}
	local, host := es[0], es[1]

	if local == "" || len(local) > maxLocalLen {
		return false
	}

	hs := strings.Split(host, dotSeparator)
	if len(hs) < 2 {
		return false
	}
	domain, ext := hs[0], hs[1]

	// Check subdomain and validate
	if len(hs) > 2 {
		if domain == "" {
			return false
		}

		for i := 1; i < len(hs)-1; i++ {
			sub := hs[i]
			if sub == "" {
				return false
			}
			domain = fmt.Sprintf("%s.%s", domain, sub)
		}

		ext = hs[len(hs)-1]
	}

	if domain == "" || len(domain) > maxDomainLen {
		return false
	}
	if ext == "" || len(ext) > maxTLDLen {
		return false
	}

	punyLocal, err := idna.ToASCII(local)
	if err != nil {
		return false
	}
	punyHost, err := idna.ToASCII(host)
	if err != nil {
		return false
	}

	if userDotRegexp.MatchString(punyLocal) || !userRegexp.MatchString(punyLocal) || !hostRegexp.MatchString(punyHost) {
		return false
	}

	return true
}
