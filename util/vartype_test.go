/// @file vartype_test.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-08-24

package util_test
import (
	"math"
	"fmt"
	"testing"
	"gitee.com/jntse/gotoolkit/util"
)

func TestLiteVarType(t *testing.T) {
	fmt.Println("============= TestLiteVarType =============")
	fmt.Println("==========")
	vbool1, vbool2, vfloat, vfloat32, vfloat64 := true, false, float64(+99.987), float32(-0.04123250), float64(+0.04123250)
	fmt.Printf("%#v\n", util.NewLiteVarType(vbool1))
	fmt.Printf("%#v\n", util.NewLiteVarType(vbool2))
	fmt.Printf("%#v\n", util.NewLiteVarType(vfloat))
	fmt.Printf("%#v\n", util.NewLiteVarType(vfloat32))
	fmt.Printf("%#v\n", util.NewLiteVarType(vfloat64))

	//
	fmt.Println("==========")
	vrune, vint, vint8, vint16, vint32, vint64 := rune(2018), int(-1), int8(2), int16(3), int32(4), int64(5)
	fmt.Printf("%#v\n", util.NewLiteVarType(vrune))
	fmt.Printf("%#v\n", util.NewLiteVarType(vint))
	fmt.Printf("%#v\n", util.NewLiteVarType(vint8))
	fmt.Printf("%#v\n", util.NewLiteVarType(vint16))
	fmt.Printf("%#v\n", util.NewLiteVarType(vint32))
	fmt.Printf("%#v\n", util.NewLiteVarType(vint64))

	//
	fmt.Println("==========")
	vuint, vuint8, vuint16, vuint32, vuint64 := uint(1), uint8(2), uint16(3), uint32(4), uint64(5)
	fmt.Printf("%#v\n", util.NewLiteVarType(vuint))
	fmt.Printf("%#v\n", util.NewLiteVarType(vuint8))
	fmt.Printf("%#v\n", util.NewLiteVarType(vuint16))
	fmt.Printf("%#v\n", util.NewLiteVarType(vuint32))
	fmt.Printf("%#v\n", util.NewLiteVarType(vuint64))

	//
	fmt.Println("==========")
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MaxInt8))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MinInt8))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MaxInt32))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MinInt32))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MaxInt64))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MinInt64))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MaxUint32))
	fmt.Printf("%#v\n", util.NewLiteVarType(uint64(math.MaxUint64)))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.SmallestNonzeroFloat32))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MaxFloat32))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.SmallestNonzeroFloat64))
	fmt.Printf("%#v\n", util.NewLiteVarType(math.MaxFloat64))

	// slice []byte
	fmt.Println("==========")
	slice1 := []byte{0xe6, 0xb1, 0x89, 0xe5, 0xad, 0x97, 0xe6, 0xb5, 0x8b, 0xe8, 0xaf, 0x95}
	fmt.Printf("%#v\n", util.NewLiteVarType(slice1))
	slice2 := "this is string"
	fmt.Printf("%#v\n", util.NewLiteVarType(slice2))
	//slice3 := []int32{1,2,3,4,5}
	//fmt.Printf("%#v\n", util.NewLiteVarType(slice3))	// panic

	//
	fmt.Println("==========")
	vstrlist := []string{"", "true", "false", "-100", "0", "+100", "99.9998", "0.0025", "anything",
						"-9223372036854775808", "18446744073709551615", "4.12325e+02", "2E+1", "2.147483647e+09", "汉字测试"}
	for _, v := range vstrlist {
		fmt.Printf("%#v\n", util.NewLiteVarType(v))
	}

	fmt.Println("============= TestLiteVarType =============")
}

func TestVarType(t *testing.T) {
	fmt.Println("============= TestVarType =============")
	//
	return
	fmt.Println("==========")
	vbool1, vbool2, vfloat, vfloat32, vfloat64 := true, false, float64(+99.987), float32(-0.04123250), float64(+0.04123250)
	fmt.Printf("%#v\n", util.NewVarType(vbool1))
	fmt.Printf("%#v\n", util.NewVarType(vbool2))
	fmt.Printf("%#v\n", util.NewVarType(vfloat))
	fmt.Printf("%#v\n", util.NewVarType(vfloat32))
	fmt.Printf("%#v\n", util.NewVarType(vfloat64))

	//
	fmt.Println("==========")
	vrune, vint, vint8, vint16, vint32, vint64 := rune(2018), int(-1), int8(2), int16(3), int32(4), int64(5)
	fmt.Printf("%#v\n", util.NewVarType(vrune))
	fmt.Printf("%#v\n", util.NewVarType(vint))
	fmt.Printf("%#v\n", util.NewVarType(vint8))
	fmt.Printf("%#v\n", util.NewVarType(vint16))
	fmt.Printf("%#v\n", util.NewVarType(vint32))
	fmt.Printf("%#v\n", util.NewVarType(vint64))

	//
	fmt.Println("==========")
	vuint, vuint8, vuint16, vuint32, vuint64 := uint(1), uint8(2), uint16(3), uint32(4), uint64(5)
	fmt.Printf("%#v\n", util.NewVarType(vuint))
	fmt.Printf("%#v\n", util.NewVarType(vuint8))
	fmt.Printf("%#v\n", util.NewVarType(vuint16))
	fmt.Printf("%#v\n", util.NewVarType(vuint32))
	fmt.Printf("%#v\n", util.NewVarType(vuint64))

	//
	fmt.Println("==========")
	fmt.Printf("%#v\n", util.NewVarType(math.MaxInt8))
	fmt.Printf("%#v\n", util.NewVarType(math.MinInt8))
	fmt.Printf("%#v\n", util.NewVarType(math.MaxInt32))
	fmt.Printf("%#v\n", util.NewVarType(math.MinInt32))
	fmt.Printf("%#v\n", util.NewVarType(math.MaxInt64))
	fmt.Printf("%#v\n", util.NewVarType(math.MinInt64))
	fmt.Printf("%#v\n", util.NewVarType(math.MaxUint32))
	fmt.Printf("%#v\n", util.NewVarType(uint64(math.MaxUint64)))
	fmt.Printf("%#v\n", util.NewVarType(math.SmallestNonzeroFloat32))
	fmt.Printf("%#v\n", util.NewVarType(math.MaxFloat32))
	fmt.Printf("%#v\n", util.NewVarType(math.SmallestNonzeroFloat64))
	fmt.Printf("%#v\n", util.NewVarType(math.MaxFloat64))

	// slice []byte
	fmt.Println("==========")
	slice1 := []byte{0xe6, 0xb1, 0x89, 0xe5, 0xad, 0x97, 0xe6, 0xb5, 0x8b, 0xe8, 0xaf, 0x95}
	fmt.Printf("%#v\n", util.NewVarType(slice1))
	slice2 := "this is string"
	fmt.Printf("%#v\n", util.NewVarType(slice2))
	//slice3 := []int32{1,2,3,4,5}
	//fmt.Printf("%#v\n", util.NewVarType(slice3))	// panic

	//
	fmt.Println("==========")
	vstrlist := []string{"", "true", "false", "-100", "0", "+100", "99.9998", "0.0025", "anything",
						"-9223372036854775808", "18446744073709551615", "4.12325e+02", "2E+1", "2.147483647e+09", "汉字测试"}
	for _, v := range vstrlist {
		fmt.Printf("%#v\n", util.NewVarType(v))
	}

	fmt.Println("============= TestVarType =============")
}

