package core

import (
	"time"
)

type CoreModel struct {
	ID int64 `bun:"id,pk,autoincrement"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}
