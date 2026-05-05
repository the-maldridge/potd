package types

import (
	"time"
)

type EscrowedToken struct {
	Host    string `gorm:"primaryKey"`
	Token   string
	Updated time.Time
}
