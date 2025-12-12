package token

import (
	"time"

	"github.com/google/uuid"
)

type Maker interface {
	CreateToken(userExternalID uuid.UUID, username string, role string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}