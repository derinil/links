package session_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/derinil/links/links/account/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCache struct{ mock.Mock }

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Put(ctx context.Context, key string, val []byte) error {
	args := m.Called(ctx, key, val)
	return args.Error(0)
}

func (m *MockCache) Invalidate(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}
func (m *MockCache) PutWithTTL(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, val, ttl)
	return args.Error(0)
}

func TestDestroy(t *testing.T) {
	testCases := []struct {
		name      string
		token     string
		err       error
		skipCache bool
	}{
		{
			name:      "empty token",
			err:       session.ErrInvalidToken,
			skipCache: true,
		},
		{
			name:  "valid destroy",
			token: "seesss",
		},
		{
			name:  "cache error",
			token: "seesss",
			err:   errors.New("haha"),
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			var (
				ctx            = context.Background()
				mockCache      = new(MockCache)
				sessionHandler = session.NewHandler(mockCache)
			)

			if !c.skipCache {
				mockCache.On("Invalidate", ctx, "session-token-"+c.token).Return(false, c.err).Once()
			}

			err := sessionHandler.Destroy(ctx, c.token)
			require.ErrorIs(t, err, c.err)

			mockCache.AssertExpectations(t)
		})
	}
}

func TestGet(t *testing.T) {
	defaultSession := session.New(uuid.New(), "handle")

	var defaultEncoded bytes.Buffer
	err := gob.NewEncoder(&defaultEncoded).Encode(defaultSession)
	require.Nil(t, err)

	testCases := []struct {
		name      string
		token     string
		encoded   []byte
		err       error
		cacheErr  error
		skipCache bool
	}{
		{
			name:      "empty token",
			skipCache: true,
			err:       session.ErrInvalidToken,
		},
		{
			name:    "valid session",
			token:   "big-token",
			encoded: defaultEncoded.Bytes(),
		},
		{
			name:  "session not found",
			token: "big-token",
			err:   session.ErrSessionNotFound,
		},
		{
			name:     "session not found",
			token:    "big-token",
			cacheErr: session.ErrSessionNotFound,
			err:      session.ErrSessionNotFound,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			var (
				ctx            = context.Background()
				mockCache      = new(MockCache)
				sessionHandler = session.NewHandler(mockCache)
			)

			if !c.skipCache {
				mockCache.On("Get", ctx, "session-token-"+c.token).
					Return(c.encoded, c.cacheErr).Once()
			}

			s, err := sessionHandler.Get(ctx, c.token)
			require.ErrorIs(t, err, c.err)

			mockCache.AssertExpectations(t)

			if c.err != nil {
				return
			}

			require.Equal(t, *defaultSession, *s)
		})
	}
}

func TestIssue(t *testing.T) {
	testCases := []struct {
		name   string
		handle string
		id     uuid.UUID
	}{
		{
			name:   "normal issue",
			handle: "handle",
			id:     uuid.New(),
		},
		{
			name:   "empty handle issue",
			handle: "",
			id:     uuid.New(),
		},
		{
			name:   "nil id issue",
			handle: "handle",
			id:     uuid.Nil,
		},
		{
			name:   "empty handle and nil id issue",
			handle: "",
			id:     uuid.Nil,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			var (
				ctx            = context.Background()
				mockCache      = new(MockCache)
				sessionHandler = session.NewHandler(mockCache)
			)

			var bs []byte

			mockCache.On("PutWithTTL", ctx,
				mock.MatchedBy(func(token string) bool {
					return strings.HasPrefix(token, "session-token-") && len(token) > 25
				}),
				mock.MatchedBy(func(b []byte) bool {
					bs = b
					return true
				}), session.Lifetime,
			).Return(nil).Once()

			s, _, err := sessionHandler.Issue(ctx, c.id, c.handle)
			require.Nil(t, err)

			mockCache.AssertExpectations(t)

			var sesh session.Session
			err = gob.NewDecoder(bytes.NewReader(bs)).Decode(&sesh)
			require.Nil(t, err)

			require.Equal(t, sesh, *s)
			require.Equal(t, c.handle, sesh.Handle)
			require.Equal(t, c.id, sesh.AccountID)
		})
	}
}
