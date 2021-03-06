/// @file vartype.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-08-24

package util
import (
	"reflect"
	"strconv"
	"strings"
	"gitee.com/jntse/gotoolkit/log"
)

type IVarType interface {
	Val() 		(interface{})
	Bool() 		(bool) 
	String() 	(string)
	Bytes()		([]byte)
	IsNil()		(bool)

	Int() 		(int)
	Int32()		(int32)
	Int64() 	(int64)

	Uint()		(uint)
	Uint32() 	(uint32)
	Uint64() 	(uint64)

	Float32() 	(float32) 
	Float64() 	(float64) 
}

// --------------------------------------------------------------------------
/// @brief 适合一次构造多次使用场景，例如成员变量
// --------------------------------------------------------------------------
type VarType struct {
	val_raw		interface{}		// rawdata
	val_bytes	[]byte
	val_string 	string
	val_int 	uint64
	val_float 	float64
	val_bool	bool
	nilflag		bool
}

//func NewVarType(valstr string)
//	return &VarType{val_string : valstr}
//}

func MakeVarTypeStringMap(set map[string]interface{}) map[string]*VarType {
	varset := make(map[string]*VarType)
	for k , v := range set {
		varset[k] = NewVarType(v)
	}
	return varset
}

func MakeVarTypeIntMap(set map[int]interface{}) map[int]*VarType {
	varset := make(map[int]*VarType)
	for k , v := range set {
		varset[k] = NewVarType(v)
	}
	return varset
}


func NewVarType(val interface{}) *VarType {
	rf, vt := reflect.ValueOf(val), &VarType{val_raw: val}

	if rf.Kind() == reflect.String {
		vt.parseString(rf.String())
	} else if rf.Kind() == reflect.Slice {	// if not []byte slice, or panic
		if rf.IsNil() == true {
			vt.nilflag = true
			return vt
		}
		vt.val_bytes = rf.Bytes()
		vt.parseString(string(vt.val_bytes))
	} else if rf.Kind() == reflect.Bool {
		vt.val_bool = rf.Bool()
		vt.val_string = strconv.FormatBool(vt.val_bool)
		vt.val_bytes = []byte(vt.val_string)
		if vt.val_bool { vt.val_float = float64(1) }
		if vt.val_bool { vt.val_int = uint64(1) }
	} else if rf.Kind() >= reflect.Int && rf.Kind() <= reflect.Int64 {
		vt.val_int = uint64(rf.Int())
		vt.val_float = float64(vt.val_int)
		vt.val_bool = (vt.val_float != 0)
		vt.val_string = strconv.FormatInt(int64(vt.val_int), 10)
		vt.val_bytes = []byte(vt.val_string)
	} else if rf.Kind() >= reflect.Uint && rf.Kind() <= reflect.Uint64 {
		vt.val_int = rf.Uint()
		vt.val_float = float64(vt.val_int)
		vt.val_bool = (vt.val_float != 0)
		vt.val_string = strconv.FormatUint(vt.val_int, 10)
		vt.val_bytes = []byte(vt.val_string)
	} else if rf.Kind() >= reflect.Float32 && rf.Kind() <= reflect.Float64 {
		vt.val_float = rf.Float()
		vt.val_int = uint64(vt.val_float)
		vt.val_bool = (vt.val_float != 0)
		if rf.Kind() == reflect.Float32 {
			vt.val_string = strconv.FormatFloat(vt.val_float, 'E', -1, 32)
		}else {
			vt.val_string = strconv.FormatFloat(vt.val_float, 'E', -1, 64)
		}
		vt.val_bytes = []byte(vt.val_string)
	}else {
		log.Error("VarType init not support variable kind=%s", rf.Type())
		return nil
	}

	return vt
}

// 解析字符串
func (vt *VarType) parseString(valstr string) {
	vt.val_string = valstr
	vt.val_bytes = []byte(vt.val_string)

	// bool
	if vt.val_string == "true" {
		vt.val_float, vt.val_int, vt.val_bool = 1, 1, true
		return
	}
	if vt.val_string == "false" {
		vt.val_float, vt.val_int, vt.val_bool = 0, 0, false
		return
	}

	// float and scientific notation，只支持的整数区间 [-(1<<63), (1<<63)]，超过受限制
	if strings.Contains(vt.val_string, ".") || strings.ContainsAny(vt.val_string, "eE") {
		vt.val_float, _ = strconv.ParseFloat(vt.val_string, 64)
		vt.val_int = uint64(vt.val_float)
		vt.val_bool = (vt.val_float != 0)
		return
	}

	// [-(1<<63), (1<<64)-1]
	if strings.HasPrefix(vt.val_string, "-") {
		valint, _   := strconv.ParseInt(vt.val_string, 10, 64)
		vt.val_int  = uint64(valint)
	}else {
		fixstring := strings.TrimLeft(vt.val_string, "+")
		vt.val_int, _ = strconv.ParseUint(fixstring, 10, 64)
	}
	vt.val_float = float64(vt.val_int)
	vt.val_bool = (vt.val_int != 0)
}

func (vt *VarType) Val() interface{} { return vt.val_raw }
func (vt *VarType) String() string { return vt.val_string }
func (vt *VarType) Bool() bool { return vt.val_bool }
func (vt *VarType) Bytes() []byte { return vt.val_bytes }
func (vt *VarType) IsNil() bool { return vt.nilflag }

func (vt *VarType) Int() (int) { return int(vt.val_int) }
func (vt *VarType) Int32() (int32) { return int32(vt.val_int) }
func (vt *VarType) Int64() (int64) { return int64(vt.val_int) }

func (vt *VarType) Uint() (uint) { return uint(vt.val_int) }
func (vt *VarType) Uint32() (uint32) { return uint32(vt.val_int) }
func (vt *VarType) Uint64() (uint64) { return uint64(vt.val_int) }

func (vt *VarType) Float32() (float32) { return float32(vt.val_float) }
func (vt *VarType) Float64() (float64) { return float64(vt.val_float) }


// --------------------------------------------------------------------------
/// @brief 轻量级可变类型
/// @brief 适合临时变量，一次构造使用
// --------------------------------------------------------------------------
type LiteVarType struct {
	val_raw interface{}
	val_string string
}

func MakeLiteVarTypeStringMap(set map[string]interface{}) map[string]*LiteVarType {
	varset := make(map[string]*LiteVarType)
	for k , v := range set {
		varset[k] = NewLiteVarType(v)
	}
	return varset
}

func MakeLiteVarTypeIntMap(set map[int]interface{}) map[int]*LiteVarType {
	varset := make(map[int]*LiteVarType)
	for k , v := range set {
		varset[k] = NewLiteVarType(v)
	}
	return varset
}

func NewLiteVarType(val interface{}) *LiteVarType {
	rf, vt := reflect.ValueOf(val), &LiteVarType{val_raw: val}
	if rf.Kind() == reflect.String {
		vt.val_string = rf.String()
	} else if rf.Kind() == reflect.Slice {
		if rf.IsNil() == true {
			return nil
		}
		vt.val_string = string(rf.Bytes())
	} else if rf.Kind() == reflect.Bool {
		vt.val_string = strconv.FormatBool(rf.Bool())
	} else if rf.Kind() >= reflect.Int && rf.Kind() <= reflect.Int64 {
		vt.val_string = strconv.FormatInt(rf.Int(), 10)
	} else if rf.Kind() >= reflect.Uint && rf.Kind() <= reflect.Uint64 {
		vt.val_string = strconv.FormatUint(rf.Uint(), 10)
	} else if rf.Kind() >= reflect.Float32 && rf.Kind() <= reflect.Float64 {
		if rf.Kind() == reflect.Float32 {
			vt.val_string = strconv.FormatFloat(rf.Float(), 'f', -1, 32)
		}else {
			vt.val_string = strconv.FormatFloat(rf.Float(), 'f', -1, 64)
		}
	}else {
		log.Error("LiteVarType init not support variable kind=%s", rf.Type())
		return nil
	}
	return vt;
}

func (vt *LiteVarType) String() string { return vt.val_string }
func (vt *LiteVarType) Val() interface{} { return vt.val_raw }
func (vt *LiteVarType) Bytes() []byte { return []byte(vt.val_string) }

func (vt *LiteVarType) IsNil() bool {
	rf := reflect.ValueOf(vt.val_raw)
	return rf.IsNil()
}

func (vt *LiteVarType) Bool() bool {
	re, _ := strconv.Atoi(vt.String())	// faster than ParseInt
	return re == 0
}

func (vt *LiteVarType) Int() (int) {
	//re, _ := strconv.ParseInt(vt.String(), 10, 32)
	re, _ := strconv.Atoi(vt.String())	// faster than ParseInt
	return int(re)
}

func (vt *LiteVarType) Int32() (int32) {
	//re, _ := strconv.ParseInt(vt.String(), 10, 32)
	re, _ := strconv.Atoi(vt.String())	// faster than ParseInt
	return int32(re)
}

func (vt *LiteVarType) Uint() (uint) { 
	//re, _ := strconv.ParseUint(vt.String(), 10, 32)
	re, _ := strconv.Atoi(vt.String())	// faster than ParseInt
	return uint(re)
}

func (vt *LiteVarType) Uint32() (uint32) {
	//re, _ := strconv.ParseUint(vt.String(), 10, 32)
	re, _ := strconv.Atoi(vt.String())	// faster than ParseInt
	return uint32(re)
}

func (vt *LiteVarType) Int64() (int64) {
	re, _ := strconv.ParseInt(vt.String(), 10, 64)
	return re
}

func (vt *LiteVarType) Uint64() (uint64) {
	re, _ := strconv.ParseUint(vt.String(), 10, 64)
	return re
}

func (vt *LiteVarType) Float32() (float32) {
	re, _ := strconv.ParseFloat(vt.String(), 32)
	return float32(re)
}

func (vt *LiteVarType) Float64() (float64) {
	re, _ := strconv.ParseFloat(vt.String(), 64)
	return re
}


