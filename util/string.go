/// @file string.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-07-09

package util
import (
	_"fmt"
	"strings"
	"strconv"
)


// 检查特殊字符
func ContainsSpecialCharacter(str string) (bool, string) {
	//sp_en := "[~!@#$%^&*()/\\|,.<>?\"'();:_+-={} "
	//sp_ch := "—…（）。，！￥；：“”‘’？、、《》"
	ch_unicode_start, ch_unicode_end := 0x4E00, 0x9FA5      // 常用中午汉字
	nu_unicode_start, nu_unicode_end := 0x0030, 0x0039      // 数字
	en_unicode_lower_start, en_unicode_lower_end := 0x0041, 0x005A  // 英文小写单词
	en_unicode_upper_start, en_unicode_upper_end := 0x0061, 0x007A  // 英文大写单词
	for _, v := range str {
		//if strings.Contains(sp_en, string(v)) == true { return true }
		//if strings.Contains(sp_ch, string(v)) == true { return true }

		if v > int32(ch_unicode_end) {
			//fmt.Println(string(v), v)
			return true , string(v)
		}

		if v <= int32(ch_unicode_start) {
			if v >= int32(nu_unicode_start) && v <= int32(nu_unicode_end) { continue }
			if v >= int32(en_unicode_lower_start) && v <= int32(en_unicode_lower_end) { continue }
			if v >= int32(en_unicode_upper_start) && v <= int32(en_unicode_upper_end) { continue }
			//fmt.Println(string(v), v)
			return true, string(v)
		}
	}

	return false, ""
}


// --------------------------------------------------------------------------
/// @brief 解析数值分割字符串
/// @brief 示例 1001-1-1;1002-2-2;2001-1-1
/// @brief 返回 []ObjSplitIntString{ []int64{1001,1,1}, []int64{1002,2,2}, []int64{2001,1,1} }
// --------------------------------------------------------------------------
type objSplitIntString struct  {		// not export for now
	Params []int64
	Rawdata string
	Error	error
}
func (obj *objSplitIntString) Len() int 	{ return len(obj.Params) }
func (obj *objSplitIntString) Raw() string 	{ return obj.Rawdata }
func (obj *objSplitIntString) Err() error 	{ return obj.Error }
func (obj *objSplitIntString) Values() []int64 { return obj.Params }
func (obj *objSplitIntString) Value(index int) int64 {
	if index > obj.Len() { 
		panic("should check objSplitIntString.Len() outside") 
	}
	return obj.Params[index]
}

func SplitIntString(src []string, sep string) []*objSplitIntString {
	objs := make([]*objSplitIntString, 0)
	for _, sub := range src {
		ssub := strings.Split(sub, sep)
		obj := &objSplitIntString{ Params:make([]int64, len(ssub)), Rawdata:sub}
		for k, v := range ssub {
			obj.Params[k], obj.Error = strconv.ParseInt(v, 10, 64)
		}
		objs = append(objs, obj)
	}
	return objs
}


