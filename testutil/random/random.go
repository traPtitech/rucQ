package random

import (
	"math"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func AlphaNumericString(t *testing.T, maxLength uint) string {
	t.Helper()

	length := rand.UintN(maxLength) + 1

	var builder strings.Builder

	for range length {
		index := rand.UintN(uint(len(charset)))

		builder.WriteByte(charset[index])
	}

	return builder.String()
}

func Bool(t *testing.T) bool {
	t.Helper()

	return rand.UintN(2) == 0
}

func PositiveInt(t *testing.T) int {
	t.Helper()

	// intの範囲に収まる正の整数を生成
	return int(rand.UintN(math.MaxInt)) + 1
}

func PtrOrNil[T any](t *testing.T, value T) *T {
	t.Helper()

	if Bool(t) {
		return nil
	}

	return &value
}

func Time(t *testing.T) time.Time {
	t.Helper()

	rand1 := rand.Int()
	rand2 := rand.Int()

	// ローカルのタイムゾーンが使われることによる不整合を回避
	return time.Now().UTC().Add(time.Duration(rand1 - rand2))
}
