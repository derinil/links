package session

import (
	"net/http"
	"time"

	"github.com/derinil/links/links/generic"
	"github.com/google/uuid"
)

type (
	Session struct {
		generic.DBStruct
		Handle    string
		AccountID uuid.UUID
		ExpiresAt time.Time
	}
)

const Lifetime = 24 * time.Hour

// Creates a session that expires in now + session.Lifetime
func New(accountID uuid.UUID, handle string) *Session {
	return &Session{
		DBStruct:  generic.NewDBStruct(),
		AccountID: accountID,
		Handle:    handle,
		ExpiresAt: time.Now().Add(Lifetime).UTC(),
	}
}

func Cookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:    CookieName,
		Value:   token,
		Expires: time.Now().Add(Lifetime),
		MaxAge:  int(Lifetime.Seconds()),
	}
}

func RemoveCookie() *http.Cookie {
	return &http.Cookie{
		Name:    CookieName,
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
		MaxAge:  0,
	}
}
