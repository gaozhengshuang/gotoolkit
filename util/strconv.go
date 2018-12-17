/// @file strconv.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-09-21

package util
import (
	"strconv"
)

func Atoi(s string) int32 {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int32(i)
}

func Atol(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func Atoq(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

func Atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func Itoa(i int32) string {
	 return strconv.FormatInt(int64(i), 10)
}

func Ltoa(l int64) string {
	 return strconv.FormatInt(l, 10)
}

func Qtoa(q uint64) string {
	 return strconv.FormatUint(q, 10)
}

func Ftoa(f float64) string {
	 return strconv.FormatFloat(f, 'G', -1, 64)
}


