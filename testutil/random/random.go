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

	const max = 2 // 0または1のいずれかが返るようにする

	return rand.UintN(max) == 0
}

func Float32(t *testing.T) float32 {
	t.Helper()

	return float32(rand.NormFloat64())
}

func Float64(t *testing.T) float64 {
	t.Helper()

	return rand.NormFloat64()
}

// intの範囲に収まる正の整数を返す
func PositiveInt(t *testing.T) int {
	t.Helper()

	return PositiveIntN(t, math.MaxInt)
}

// [1, n]の整数を返す
func PositiveIntN(t *testing.T, n uint) int {
	t.Helper()

	return int(rand.UintN(n)) + 1
}

func PtrOrNil[T any](t *testing.T, value T) *T {
	t.Helper()

	if Bool(t) {
		return nil
	}

	return &value
}

func SelectFrom[T any](t *testing.T, items ...T) T {
	t.Helper()

	index := rand.UintN(uint(len(items)))

	return items[index]
}

func Time(t *testing.T) time.Time {
	t.Helper()

	rand1 := rand.Int()
	rand2 := rand.Int()

	// ローカルのタイムゾーンが使われることによる不整合を回避
	return time.Now().UTC().Add(time.Duration(rand1 - rand2))
}
