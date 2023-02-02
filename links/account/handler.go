package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/derinil/links/links/generic"
	"github.com/google/uuid"
)

type (
	Handler interface {
		Get(ctx context.Context, cmd *GetCmd) (*Account, error)
		Create(ctx context.Context, cmd *CreateCmd) (*Account, error)
		Update(ctx context.Context, cmd *UpdateCmd) (*Account, error)
	}

	HandlerImpl struct {
		reader Reader
		writer Writer
	}

	Reader interface {
		Get(ctx context.Context, cmd *GetCmd) (*Account, error)
	}

	Writer interface {
		SaveAccount(ctx context.Context, a *Account) error
	}

	CreateCmd struct {
		Name     string
		Handle   string
		Password string
	}

	GetCmd struct {
		ID      uuid.UUID
		Handle  string
		Shallow bool
	}

	UpdateCmd struct {
		AccountID uuid.UUID
		Name      string
		Handle    string
		CSS       string
		Links     []LinkScaffold
	}

	LinkScaffold struct {
		Title string
		Link  string
	}
)

var (
	ErrAccountNotFound = generic.NewWebError(http.StatusNotFound, "account_not_found", "Account not found")
	ErrHandleTaken     = generic.NewWebError(http.StatusBadRequest, "handle_taken", "Handle is already taken")
)

var _ Handler = (*HandlerImpl)(nil)

func NewHandler(reader Reader, writer Writer) *HandlerImpl {
	return &HandlerImpl{reader: reader, writer: writer}
}

func (s *HandlerImpl) Create(ctx context.Context, cmd *CreateCmd) (*Account, error) {
	a := New(cmd.Name, cmd.Handle, cmd.Password)

	a.Sanitize()
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate account: %w", err)
	}

	ea, err := s.reader.Get(ctx, &GetCmd{Handle: a.Handle})
	if err != nil {
		return nil, fmt.Errorf("failed to check if handle is taken: %w", err)
	}

	if ea != nil {
		return nil, ErrHandleTaken
	}

	if err := s.writer.SaveAccount(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return a, nil
}

func (s *HandlerImpl) Update(ctx context.Context, cmd *UpdateCmd) (*Account, error) {
	a, err := s.reader.Get(ctx, &GetCmd{ID: cmd.AccountID})
	if err != nil {
		return nil, fmt.Errorf("failed to get account by id: %w", err)
	}

	if a == nil {
		return nil, ErrAccountNotFound
	}

	if cmd.Name != "" {
		a.Name = cmd.Name
	}
	if cmd.Handle != "" {
		a.Handle = cmd.Handle
	}
	if cmd.CSS != "" {
		a.CSS = cmd.CSS
	}

	a.Sanitize()
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate account: %w", err)
	}

	oldLinks := make(map[string]*Link, len(a.Links))
	for i := range a.Links {
		l := &a.Links[i]
		oldLinks[l.Link] = l
	}

	for i := range cmd.Links {
		l := &cmd.Links[i]

		nl := NewLink(a.ID, l.Title, l.Link, i)

		nl.Sanitize()
		if err := nl.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate link: %w", err)
		}

		ol, ok := oldLinks[nl.Link]
		if ok {
			ol.Title = nl.Title
			continue
		}

		a.Links = append(a.Links, *nl)
	}

	ea, err := s.reader.Get(ctx, &GetCmd{Handle: a.Handle})
	if err != nil {
		return nil, fmt.Errorf("failed to get account by new handle: %w", err)
	}

	if ea != nil && ea.ID != a.ID {
		return nil, ErrHandleTaken
	}

	if err := s.writer.SaveAccount(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}

	return a, nil
}

func (s *HandlerImpl) Get(ctx context.Context, cmd *GetCmd) (*Account, error) {
	a, err := s.reader.Get(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by id: %w", err)
	}

	if a == nil {
		return nil, ErrAccountNotFound
	}

	return a, nil
}
