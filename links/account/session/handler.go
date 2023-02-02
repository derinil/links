package session

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/derinil/links/links/cache"
	"github.com/derinil/links/links/generic"
	"github.com/google/uuid"
)

type (
	Handler interface {
		Destroy(ctx context.Context, token string) error
		Get(ctx context.Context, token string) (*Session, error)
		Issue(ctx context.Context, accountID uuid.UUID, handle string) (*Session, string, error)
	}

	HandlerImpl struct {
		cache cache.Cache
	}

	CtxKey string
)

const (
	CookieName string = "session"
)

var (
	ErrInvalidToken     = generic.NewWebError(http.StatusBadRequest, "token_invalid", "Session token is invalid")
	ErrSessionNotFound  = generic.NewWebError(http.StatusUnauthorized, "session_not_found", "Session is invalid")
	ErrNotAuthenticated = generic.NewWebError(http.StatusUnauthorized, "not_authorized", "You are not logged in")
)

var _ Handler = (*HandlerImpl)(nil)

func NewHandler(cache cache.Cache) *HandlerImpl {
	return &HandlerImpl{cache: cache}
}

func (s *HandlerImpl) Get(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	sea, err := s.cache.Get(ctx, cacheKey(token))
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if len(sea) == 0 {
		return nil, ErrSessionNotFound
	}

	var se Session
	if err := gob.NewDecoder(bytes.NewReader(sea)).Decode(&se); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	return &se, nil
}

func (s *HandlerImpl) Destroy(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidToken
	}

	_, err := s.cache.Invalidate(ctx, cacheKey(token))
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

func (s *HandlerImpl) Issue(ctx context.Context, accountID uuid.UUID, handle string) (*Session, string, error) {
	se := New(accountID, handle)

	t, err := s.createToken(se.ID[:])
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session token: %w", err)
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(se); err != nil {
		return nil, "", fmt.Errorf("failed to encode session: %w", err)
	}

	if err = s.cache.PutWithTTL(ctx, cacheKey(t), b.Bytes(), Lifetime); err != nil {
		return nil, "", fmt.Errorf("failed to cache session: %w", err)
	}

	return se, t, nil
}

func (s *HandlerImpl) createToken(id []byte) (string, error) {
	h := sha256.New()

	_, err := h.Write(id)
	if err != nil {
		return "", fmt.Errorf("failed to write session id: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func cacheKey(t string) string {
	return "session-token-" + t
}
