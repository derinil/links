package account

import (
	"fmt"
	"strings"

	"github.com/derinil/links/links/generic"
)

type Account struct {
	generic.DBStruct
	Name     string `validate:"max=128" db:"name"`
	Handle   string `validate:"handle" db:"handle"`
	Password string `validate:"max=5000,css" db:"password"`
	CSS      string `validate:"css" db:"css"`
	Avi      []byte `db:"avi"`
	Links    []Link `db:"-"`
}

func New(name, handle, password string) *Account {
	return &Account{
		DBStruct: generic.NewDBStruct(),
		Name:     name,
		Handle:   handle,
		Password: password,
	}
}

func (a *Account) Sanitize() {
	a.CSS = strings.TrimSpace(a.CSS)
	a.Name = strings.TrimSpace(a.Name)
	a.Handle = strings.ToLower(strings.TrimSpace(a.Handle))
}

func (a *Account) Validate() error {
	if err := generic.Validator.Struct(a); err != nil {
		return fmt.Errorf("failed to validate account: %w", err)
	}

	return nil
}

func (a *Account) BeforeSave() error {
	for i := range a.Links {
		a.Links[i].AccountID = a.ID
		a.Links[i].Index = i
	}

	a.Sanitize()
	if err := a.Validate(); err != nil {
		return fmt.Errorf("failed to validate account: %w", err)
	}

	a.SetUpdatedAt()

	return nil
}

func (a *Account) AfterLoad() error {
	return nil
}
