/// @file bytes_test.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-03-18

package util
import (
	"testing"
	"fmt"
)

func CheckEqual(t *testing.T, val interface{}, exp interface{}) {
	if val != exp {
		t.Errorf("Expected %v, got %v.", exp, val)
	}
}

func TestBoolToBytes(t *testing.T) {
	buf := BoolToBytes(false)
	v := BytesToBool(buf)
	CheckEqual(t, v, false)
}

func TestInt32ToBytes(t *testing.T) {
	n := int32(32)
	buf := Int32ToBytes(n)
	v := BytesToInt32(buf)
	CheckEqual(t, v, n)
}

func TestInt64ToBytes(t *testing.T) {
	n := int64(32)
	buf := Int64ToBytes(n)
	v := BytesToInt64(buf)
	CheckEqual(t, v, n)
}

func TestFloat32ToByte(t *testing.T) {
	n := float32(12345.6789)
	buf := Float32ToBytes(n)
	v := BytesToFloat32(buf)
	CheckEqual(t, v, n)
}

func TestFloat64ToByte(t *testing.T) {
	n := float64(123456789.987654321)
	buf := Float64ToBytes(n)
	fmt.Println(buf)
	v := BytesToFloat64(buf)
	fmt.Println(v)
	CheckEqual(t, v, n)
}


