package account

import (
	"fmt"
	"strings"

	"github.com/derinil/links/links/generic"
	"github.com/google/uuid"
)

type Link struct {
	generic.DBStruct
	AccountID uuid.UUID `db:"account_id"`
	Title     string    `validate:"min=1,max=128" db:"title"`
	Link      string    `validate:"url" db:"link"`
	Favicon   []byte    `db:"favicon"`
	Index     int       `db:"index"`
}

func NewLink(
	accountID uuid.UUID,
	title string,
	link string,
	index int,
) *Link {
	return &Link{
		DBStruct:  generic.NewDBStruct(),
		AccountID: accountID,
		Title:     title,
		Link:      link,
		Index:     index,
	}
}

func (l *Link) Sanitize() {
	l.Title = strings.TrimSpace(l.Title)
	l.Link = strings.TrimSpace(l.Link)
}

func (l *Link) Validate() error {
	if err := generic.Validator.Struct(l); err != nil {
		return fmt.Errorf("failed to validate link: %w", err)
	}

	return nil
}

func (l *Link) BeforeSave() error {
	l.Sanitize()
	if err := l.Validate(); err != nil {
		return fmt.Errorf("failed to validate account: %w", err)
	}

	l.SetUpdatedAt()

	return nil
}

func (l *Link) AfterLoad() error {
	return nil
}
