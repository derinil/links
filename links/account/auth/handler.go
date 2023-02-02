package auth

import (
	"context"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/account/session"
	"github.com/derinil/links/links/generic"
)

type (
	Handler interface {
		Handle(ctx context.Context, cmd *AuthCmd) (*Auth, error)
	}

	HandlerImpl struct {
		methods map[Method]Auther
	}

	Auth struct {
		SessionToken string
		Account      *account.Account
		Session      *session.Session
	}

	AuthCmd struct {
		generic.DBStruct
		Method Method
		Cmd    any
	}

	Method string

	Auther interface {
		Method() Method
		Handle(ctx context.Context, cmd any) (*Auth, error)
	}
)

const (
	Register Method = "register"
	Logout   Method = "logout"
	Login    Method = "login"
)

var _ Handler = (*HandlerImpl)(nil)

func NewHandler(hs ...Auther) *HandlerImpl {
	m := make(map[Method]Auther, len(hs))
	for i := range hs {
		m[hs[i].Method()] = hs[i]
	}

	return &HandlerImpl{methods: m}
}

func (s *HandlerImpl) Handle(ctx context.Context, cmd *AuthCmd) (*Auth, error) {
	return s.methods[cmd.Method].Handle(ctx, cmd.Cmd)
}
