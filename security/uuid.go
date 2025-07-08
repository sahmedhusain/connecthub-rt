package security

import (
	"log"

	"github.com/google/uuid"
)

func GenerateToken() (uuid.UUID, error) {
	log.Printf("[DEBUG] Generating new UUID token")

	token, err := uuid.NewRandom()
	if err != nil {
		log.Printf("[ERROR] Failed to generate UUID token: %v", err)
		return uuid.Nil, err
	}

	maskedToken := maskUUID(token.String())
	log.Printf("[INFO] Generated new UUID token: %s", maskedToken)

	return token, nil
}

func maskUUID(uuidStr string) string {
	if len(uuidStr) < 12 {
		return uuidStr
	}

	return uuidStr[:4] + "********-****-****-****-********" + uuidStr[len(uuidStr)-4:]
}
