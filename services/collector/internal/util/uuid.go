package util

import (
	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
