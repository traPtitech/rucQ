package random

import (
	"math"
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

func PositiveInt(t *testing.T) int {
	t.Helper()

	// intの範囲に収まる正の整数を生成
	return int(rand.UintN(math.MaxInt)) + 1
}

func Time(t *testing.T) time.Time {
	t.Helper()

	rand1 := rand.Int()
	rand2 := rand.Int()

	// ローカルのタイムゾーンが使われることによる不整合を回避
	return time.Now().UTC().Add(time.Duration(rand1 - rand2))
}
