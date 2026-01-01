package utils

import "math"

// SafeIntToInt32 はintをint32に安全に変換する
// オーバーフローする場合はmath.MaxInt32/MinInt32を返す
func SafeIntToInt32(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}
