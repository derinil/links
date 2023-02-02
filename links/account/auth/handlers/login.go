package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/account/auth"
	"github.com/derinil/links/links/account/session"
	"github.com/derinil/links/links/crypto"
	"github.com/derinil/links/links/generic"
)

type LoginCmd struct {
	Handle   string
	Password string
}

var ErrLoginInvalid = generic.NewWebError(http.StatusBadRequest, "login_invalid", "Login failed")

func LoginHandler(
	accountHandler account.Handler,
	sessionHandler session.Handler,
) *Handler {
	return &Handler{
		method: auth.Login,
		handle: func(ctx context.Context, cmda any) (*auth.Auth, error) {
			cmd := cmda.(*LoginCmd)

			a, err := accountHandler.Get(ctx, &account.GetCmd{Handle: cmd.Handle})
			if err != nil {
				return nil, fmt.Errorf("failed to get account: %w", err)
			}

			if a == nil {
				return nil, ErrLoginInvalid
			}

			ok, err := crypto.CompareSha256(cmd.Password, a.Handle, a.Password)
			if err != nil {
				return nil, fmt.Errorf("failed to compare passwords: %w", err)
			}

			if !ok {
				return nil, ErrLoginInvalid
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
