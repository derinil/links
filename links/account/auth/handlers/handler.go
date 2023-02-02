package handlers

import (
	"context"

	"github.com/derinil/links/links/account/auth"
)

type Handler struct {
	method auth.Method
	handle func(ctx context.Context, cmd any) (*auth.Auth, error)
}

var _ auth.Auther = (*Handler)(nil)

func (s *Handler) Method() auth.Method {
	return s.method
}

func (s Handler) Handle(ctx context.Context, cmd any) (*auth.Auth, error) {
	return s.handle(ctx, cmd)
}
