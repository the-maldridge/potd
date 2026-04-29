package types

import (
	"time"
)

type EscrowedToken struct {
	ID uint

	Token   string
	Host    string
	Updated time.Time
}
