package csrf

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/derinil/links/links/crypto"
)

type (
	Handler interface {
		CreateCookie() (string, error)
		CreateToken(cookie string) (string, error)
		ValidateCookie(cookie string) (bool, error)
		ValidateToken(token, cookie string) (bool, error)
		ValidateCookieToken(cookie, token string) (bool, error)
	}

	HandlerImpl struct {
		key []byte
	}
)

/*
	CSRF Prevention System:
		- Initially we create a seed cookie that'll persist for a session
			- Seed cookie will be the HMAC hash of a timestamp and a random hex string
				that way no two requests will have the same cookie. We will also hex encode
				the whole cookie to make it look more presentable. So the format will be:
				<hex string> -> <hex encoded hmac:seed:timestamp> -> sha256 hmac of seed and timestamp
		- We then use the same flow of creating a token but instead of a random hex string we use the cookie
			itself along with a timestamp once again.
*/

var _ Handler = (*HandlerImpl)(nil)

func NewHandler(key []byte) *HandlerImpl {
	if len(key) == 0 {
		panic("empty csrf key")
	}

	return &HandlerImpl{key: key}
}

func (s *HandlerImpl) CreateCookie() (string, error) {
	hs, err := crypto.ReadHex(18)
	if err != nil {
		return "", fmt.Errorf("failed to read hex string: %w", err)
	}

	seconds := strconv.FormatInt(time.Now().Unix(), 10)

	h, err := s.calculateHMAC(seconds, hs)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hmac: %w", err)
	}

	return fmt.Sprintf("%s:%s:%s", hex.EncodeToString(h), seconds, hs), nil
}

func (s *HandlerImpl) CreateToken(cookie string) (string, error) {
	seconds := strconv.FormatInt(time.Now().Unix(), 10)

	h, err := s.calculateHMAC(seconds, cookie)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hmac: %w", err)
	}

	return fmt.Sprintf("%s:%s", hex.EncodeToString(h), seconds), nil
}

func (s *HandlerImpl) ValidateCookie(cookie string) (bool, error) {
	args := strings.Split(cookie, ":")
	if len(args) != 3 {
		return false, fmt.Errorf("invalid args len: %d", len(args))
	}

	var (
		hh      = args[0]
		seconds = args[1]
		hs      = args[2]
	)

	nh, err := s.calculateHMAC(seconds, hs)
	if err != nil {
		return false, fmt.Errorf("failed to calculate hmac: %w", err)
	}

	h, err := hex.DecodeString(hh)
	if err != nil {
		return false, fmt.Errorf("failed to decode hex hmac: %w", err)
	}

	return hmac.Equal(h, nh), nil
}

func (s *HandlerImpl) ValidateToken(token, cookie string) (bool, error) {
	args := strings.Split(token, ":")
	if len(args) != 2 {
		return false, fmt.Errorf("invalid args len: %d", len(args))
	}

	var (
		hh      = args[0]
		seconds = args[1]
	)

	nh, err := s.calculateHMAC(seconds, cookie)
	if err != nil {
		return false, fmt.Errorf("failed to calculate hmac: %w", err)
	}

	h, err := hex.DecodeString(hh)
	if err != nil {
		return false, fmt.Errorf("failed to decode hex hmac: %w", err)
	}

	return hmac.Equal(h, nh), nil
}

func (s *HandlerImpl) ValidateCookieToken(cookie, token string) (bool, error) {
	if cookie == "" || token == "" {
		return false, nil
	}

	ok, err := s.ValidateCookie(cookie)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	ok, err = s.ValidateToken(token, cookie)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s *HandlerImpl) calculateHMAC(args ...string) ([]byte, error) {
	h := hmac.New(sha256.New, s.key)

	for _, a := range args {
		if _, err := fmt.Fprint(h, a); err != nil {
			return nil, fmt.Errorf("failed to write arg: %w", err)
		}
	}

	return h.Sum(nil), nil
}
