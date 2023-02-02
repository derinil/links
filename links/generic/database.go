package generic

import (
	"time"

	"github.com/google/uuid"
)

type DBStruct struct {
	ID         uuid.UUID `db:"id"`
	InsertedAt time.Time `db:"inserted_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func NewDBStruct() DBStruct {
	return DBStruct{
		ID:         uuid.New(),
		InsertedAt: time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}

func (d *DBStruct) SetUpdatedAt() {
	d.UpdatedAt = time.Now().UTC()
}
