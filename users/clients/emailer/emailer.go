// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/users/clients"
)

var _ clients.Emailer = (*emailer)(nil)

const (
	passwordResetSubject = "Password Reset Request"
	inivitationSubject   = "Invitation"
	invitationBody       = `
We are excited to invite you to join our platform at {{HOST}}! To get started, please click on the invitation link below:

{{URL}}

Once you click the link, you'll be able to create your account and access all the amazing features we have to offer.

If you didn't expect this invitation, please disregard this message.
`
	resetBody = `
Dear {{USER}},

We have received a request to reset your password for your account on {{HOST}}. To proceed with resetting your password, please click on the link below:

{{URL}}

If you did not initiate this request, please disregard this message and your password will remain unchanged.
`
)

type emailer struct {
	resetURL     string
	invitaionURL string
	agent        *email.Agent
}

// New creates new emailer utility.
func New(resetURL, invitationURL string, c *email.Config) (clients.Emailer, error) {
	e, err := email.New(c)

	return &emailer{
		resetURL:     resetURL,
		invitaionURL: invitationURL,
		agent:        e,
	}, err
}

func (e *emailer) SendPasswordReset(to, host, userName, token string) error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)

	var content = strings.Replace(resetBody, "{{HOST}}", host, 1)
	content = strings.Replace(content, "{{URL}}", url, 1)
	content = strings.Replace(content, "{{USER}}", userName, 1)

	return e.agent.Send([]string{to}, "", passwordResetSubject, "", host, userName, content, "")
}

func (e *emailer) SendInvitation(to, host, token string) error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.invitaionURL, token)

	var content = strings.Replace(invitationBody, "{{HOST}}", host, 1)
	content = strings.Replace(content, "{{URL}}", url, 1)

	return e.agent.Send([]string{to}, "", inivitationSubject, "", host, "", content, "")
}
