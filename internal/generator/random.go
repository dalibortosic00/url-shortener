package generator

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type RandomGenerator struct {
	length int
}

func NewRandomGenerator(length int) *RandomGenerator {
	return &RandomGenerator{length: length}
}

func (g *RandomGenerator) Generate() string {
	result := make([]byte, g.length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}
