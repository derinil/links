package handlers

import (
	"context"
	"fmt"

	"github.com/derinil/links/links/account/auth"
	"github.com/derinil/links/links/account/session"
)

type LogoutCmd struct {
	SessionToken string
}

func LogoutHandler(sessionHandler session.Handler) *Handler {
	return &Handler{
		method: auth.Logout,
		handle: func(ctx context.Context, cmda any) (*auth.Auth, error) {
			cmd := cmda.(*LogoutCmd)

			if err := sessionHandler.Destroy(ctx, cmd.SessionToken); err != nil {
				return nil, fmt.Errorf("failed to destroy session: %w", err)
			}

			return &auth.Auth{}, nil
		},
	}
}
