/// @file math.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-11-15

package util

import (
	"math"
)

// --------------------------------------------------------------------------
/// @brief Returns the largest of x and y. If both are equivalent, x is returned
// --------------------------------------------------------------------------
func Max(x, y float64) float64 {
	return math.Max(x, y)
}

func MaxFloat32(x, y float32) float32 {
	return float32(math.Max(float64(x), float64(y)))
}


func MaxInt(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func MaxUint(x, y uint) uint {
	if x >= y {
		return x
	}
	return y
}


func MaxInt32(x, y int32) int32 {
	if x >= y {
		return x
	}
	return y
}

func MaxUint32(x, y uint32) uint32 {
	if x >= y {
		return x
	}
	return y
}

func MaxInt64(x, y int64) int64 {
	if x >= y {
		return x
	}
	return y
}

func MaxUint64(x, y uint64) uint64 {
	if x >= y {
		return x
	}
	return y
}

// --------------------------------------------------------------------------
/// @brief Returns the smallest of x and y. If both are equivalent, x is returned.
// --------------------------------------------------------------------------
func Min(x, y float64) float64 {
	return math.Min(x, y)
}

func MinFloat32(x, y float32) float32 {
	return float32(math.Min(float64(x), float64(y)))
}

func MinInt(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func MinUint(x, y uint) uint {
	if x <= y {
		return x
	}
	return y
}


func MinInt32(x, y int32) int32 {
	if x <= y {
		return x
	}
	return y
}

func MinUint32(x, y uint32) uint32 {
	if x <= y {
		return x
	}
	return y
}


func MinInt64(x, y int64) int64 {
	if x <= y {
		return x
	}
	return y
}

func MinUint64(x, y uint64) uint64 {
	if x <= y {
		return x
	}
	return y
}

// --------------------------------------------------------------------------
/// @brief 获得能够被base整除的最接近x的数
/// @brief fun(12345, 5)    -> 12345
/// @brief fun(12345, 10)   -> 12340
/// @brief fun(12345, 100)  -> 12300
// --------------------------------------------------------------------------
func FloorByBase(x, base int64) int64 {
	if base == 0 {
		return x
	}
	rem := x % base
	return x - rem
}


