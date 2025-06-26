package random

import (
	"math/rand/v2"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func AlphaNumericString(t *testing.T, maxLength uint) string {
	t.Helper()

	length := rand.UintN(maxLength) + 1

	var result string

	for range length {
		index := rand.UintN(uint(len(charset)))
		result += string(charset[index])
	}

	return result
}

func PositiveNumber(t *testing.T) uint {
	t.Helper()

	return rand.Uint() + 1
}

func Time(t *testing.T) time.Time {
	t.Helper()

	rand1 := rand.Int()
	rand2 := rand.Int()

	return time.Now().Add(time.Duration(rand1 - rand2))
}
