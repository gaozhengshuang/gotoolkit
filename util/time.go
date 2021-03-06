/// @file time.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import (
	"time"
)

// 常量
const (
	MinuteSec = 60
	HourSec = 3600
	DaySec = 86400
)

// CURTIME 秒
func CURTIME() int64 {
	return time.Now().Unix()
}

// CURTIMENS 纳秒
func CURTIMENS() int64 {
	return time.Now().UnixNano()
}

// CURTIMEUS 微秒
func CURTIMEUS() int64 {
	return time.Now().UnixNano() / 1000
}

// CURTIMEMS 毫秒
func CURTIMEMS() int64 {
	return time.Now().UnixNano() / 1000000
}

// FloorIntClock 获得当前时间上一个整点或者当前整点，秒
func FloorIntClock(t int64) int64 {
	ft := int64(t / 3600) * 3600
    return ft
}

// CeilIntClock 获得当前时间下一个整点，秒
func CeilIntClock(t int64) int64 {
	ct := int64(t / 3600) * 3600 + 3600
    return ct
}

// 格式化时间
func ReadableTime(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

// IsSameDay 是否同一天
func IsSameDay(tm1 int64, tm2 int64) bool{
    gotime1, gotime2:= time.Unix(tm1, 0), time.Unix(tm2, 0)
    if gotime1.Year() == gotime2.Year() && gotime1.YearDay() == gotime2.YearDay(){
        return true
    }
    return false         
}

// IsSameWeek 是否同一周
func IsSameWeek(tm1 int64, tm2 int64) bool{
	//if GetWeekStart(tm1) == GetWeekStart(tm2) {
	//	return true
	//}
	y1, w1 := time.Unix(tm1, 0).ISOWeek()
	y2, w2 := time.Unix(tm2, 0).ISOWeek()
	if y1 == y2 && w1 == w2 {
		return true
	}
	return false
}

// IsSameMonth 是否同一周
func IsSameMonth(tm1 int64, tm2 int64) bool{
    gotime1, gotime2 := time.Unix(tm1, 0), time.Unix(tm2, 0)
	if gotime1.Year() == gotime2.Year() && gotime1.Month() == gotime2.Month() { 
		return true
	}
	return false
}

// IsNextDay 是否相邻天
func IsNextDay(tm1 int64, tm2 int64) bool{
    if tm1 > tm2{
        tmptime := tm1                 
        tm1 = tm2
        tm2 = tmptime
    }
    if !IsSameDay(tm1, tm2) && tm2 -tm1 < DaySec * 2 {
        return true
    }
    return false
}

// GetDayStart 今日零点
func GetDayStart() int64 {
	return GetTimeDayStart(time.Now().Unix())
}

// GetTimeDayStart 指定时间(ms)零点
func GetTimeDayStart(t int64) int64 {
	tm := time.Unix(t, 0)
	y, m, d, lo := tm.Year(), tm.Month(), tm.Day(), tm.Location()
	return time.Date(y, m, d, 0, 0, 0, 0, lo).Unix()
}

// GetWeekStart 周开始秒数(星期一的零点)
func GetWeekStart(t int64) int64 {
	tm := time.Unix(t, 0)
	y, m, d, lo, wday := tm.Year(), tm.Month(), tm.Day(), tm.Location(), int64(tm.Weekday())
	daystart := time.Date(y, m, d, 0, 0, 0, 0, lo).Unix()

	if wday == 0 { wday = 7 }
	weekstart := daystart - (wday - 1) * DaySec	// 星期天是0
	return weekstart
}

// --------------------------------------------------------------------------
/// @brief StatFunctionTimeConsume 函数时间消耗统计
// --------------------------------------------------------------------------
type StatFunctionTimeConsume struct {
	capcity	int32		// 初始容量
	wpos 	int32		// 记录偏移
	rpos 	int32		// 读取偏移
	records []int64		// 记录
}

func (f *StatFunctionTimeConsume) Init(n int32) {
	f.capcity 	= n
	f.wpos 		= 0
	f.rpos  	= 0
	f.records = make([]int64, n, n)
}

func (f *StatFunctionTimeConsume) Write(t int64) bool {
	if f.wpos >= f.capcity {
		return false
	}

	f.records[f.wpos] = t
	f.wpos++
	return true
}

func (f *StatFunctionTimeConsume) Read() int64 {
	if f.rpos + 2 > f.wpos {
		return 0
	}

	diff := f.records[f.rpos+1] - f.records[f.rpos]
	f.rpos++
	return diff
}

func (f *StatFunctionTimeConsume) Total() int64 {
	if f.wpos <= 0 {
		return 0
	}
	return f.records[f.wpos-1] - f.records[0]
}

func (f *StatFunctionTimeConsume) Reset() {
	f.wpos = 0
	f.rpos = 0
}

