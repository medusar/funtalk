package common

import (
	"github.com/google/uuid"
)

func NewId() uuid.UUID {
	return uuid.New()
}

func NewIdStr() string {
	return string(NewId().ID())
}
