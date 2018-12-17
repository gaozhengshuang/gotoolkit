/// @file cmdline.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import (
	"fmt"
	"reflect"
	"flag"
	"strings"
)

// --------------------------------------------------------------------------
/// @brief 命令行解析
/// @brief 支持5个常用类型 string, float64, int, int64，bool 可扩展
/// @brief 特别注意: 传入的object结构体字段都必须是大写，reflect.Value.SetXXX() 不能操作未导出的字段
/// @return 
// --------------------------------------------------------------------------
func DoCmdLineParse(object interface{}) {
	newCmdLineParser(object).unmarshal()
}

type CmdLineParser struct {
	object interface{}
}

func newCmdLineParser(object interface{}) *CmdLineParser {
	ptr := &CmdLineParser{ object : object }
	return ptr
}

func (t* CmdLineParser) unmarshal ()  {
	vf := reflect.ValueOf(t.object)
	if vf.Kind() != reflect.Ptr { panic("'object' must be a pointer of struct") }
	s := vf.Elem()
	typeOfT := s.Type()
	values := make([]interface{},0)

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fname := strings.ToLower(typeOfT.Field(i).Name)	// 使用小写注册
		//fmt.Printf("[%d]: %s %s = %v\n", i, fname, f.Type(), f.Interface())
		switch f.Kind() {
		case reflect.String:
			str := flag.String(fname, "", "input -" + fname)
			values = append(values, str)
		case reflect.Float64:
			fval := flag.Float64(fname, 0, "input -" + fname)
			values = append(values, fval)
		case reflect.Int:
			ival := flag.Int(fname, 0, "input -" + fname)
			values = append(values, ival)
		case reflect.Int64:
			ival := flag.Int64(fname, 0, "input -" + fname)
			values = append(values, ival)
		case reflect.Bool:
			bval := flag.Bool(fname, false, "input -" + fname)
			values = append(values, bval)
		default:
			panic(fmt.Sprintf("cmdline argument not support kind=%s", f.Type()))
		}
	}

	flag.Parse()	// be carefull

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ptr := values[i]
		switch f.Kind() {
		case reflect.String:
			val := *ptr.(*string)
			f.SetString(val)
		case reflect.Float64:
			val := *ptr.(*float64)
			f.SetFloat(float64(val))
		case reflect.Int:
			val := *ptr.(*int)
			f.SetInt(int64(val))
		case reflect.Int64:
			val := *ptr.(*int64)
			f.SetInt(val)
		case reflect.Bool:
			val := *ptr.(*bool)
			f.SetBool(val)
		default:
			panic(fmt.Sprintf("not support kind=%s", f.Type()))
		}
	}
}


