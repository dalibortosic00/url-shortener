package util

import (
	"github.com/google/uuid"
)

func GenerateUserID() string {
	return uuid.NewString()
}
