package handlers

import (
	"context"
	"fmt"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/account/auth"
	"github.com/derinil/links/links/account/session"
	"github.com/derinil/links/links/crypto"
)

type RegistrationCmd struct {
	Name     string
	Handle   string
	Password string
}

func RegistrationHandler(
	accountHandler account.Handler,
	sessionHandler session.Handler,
) *Handler {
	return &Handler{
		method: auth.Register,
		handle: func(ctx context.Context, cmda any) (*auth.Auth, error) {
			cmd := cmda.(*RegistrationCmd)

			pw, err := crypto.Sha256(cmd.Password, cmd.Handle)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}

			a, err := accountHandler.Create(ctx, &account.CreateCmd{
				Password: pw,
				Name:     cmd.Name,
				Handle:   cmd.Handle,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create account: %w", err)
			}

			s, t, err := sessionHandler.Issue(ctx, a.ID, a.Handle)
			if err != nil {
				return nil, fmt.Errorf("failed to issue session: %w", err)
			}

			return &auth.Auth{
				Account:      a,
				SessionToken: t,
				Session:      s,
			}, nil
		},
	}
}
