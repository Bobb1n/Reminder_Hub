package util

import (
	"crypto/rand"
	"fmt"

	"github.com/google/uuid"
)

func GenerateUUID() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {

		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
	}
	return uuid.String(), nil
}
