/// @file define.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import "fmt"
import "strings"
import "strconv"
import "runtime"
import "time"
import "math/rand"
//import "os"

var Pf = fmt.Printf
var Pln = fmt.Println

type DWORD uint32
type QWORD uint64
type SDWORD int32
type SQWORD int64

func init() {
	rand.Seed(time.Now().Unix())
}

// 获取goroutine id，效率较低只在debug使用
func GetRoutineID() int {
	var buf [128]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// uuid生成器，使用闭包
var UUID func() int64 = UUIDGenerator()
func UUIDGenerator() func() int64 {
	var counter uint16 = 0
	return func() int64{
		tm := CURTIME()
		counter++
		uuid := tm << 32 | int64(counter)
		return uuid
	}
}

// uuid生成器，使用全局变量
var _uuid_counter_ uint16 = 0
func UUID2() int64 {
	tm := int64(CURTIME())
	_uuid_counter_++
	uuid := tm << 32 | int64(_uuid_counter_)
	return uuid
}

// uuid生成器，使用纳秒做uuid(连续两次调用间隔在10微秒左右,不会出现相同)
func UUID3() int64 {
	uuid := int64(CURTIMENS())
	return uuid
}


