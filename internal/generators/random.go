package generators

import (
	"crypto/rand"
	"fmt"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type RandomGenerator struct{}

func NewRandomGenerator() *RandomGenerator {
	return &RandomGenerator{}
}

func (g *RandomGenerator) Generate(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid key length")
	}

	result := make([]byte, length)

	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("entropy source failure: %w", err)
	}

	for i := range length {
		result[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return string(result), nil
}
