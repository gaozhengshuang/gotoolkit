/// @file paniclog.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import "runtime/debug"
import "gitee.com/jntse/gotoolkit/log"
import "time"


// --------------------------------------------------------------------------
/// @brief 1. recover内建函数被用于从 panic 或 错误场景中恢复
/// @brief 2. 让程序可以从 panicking 重新获得控制权，停止终止过程进而恢复正常执行，
/// @brief 3. recover后逻辑并不会恢复到panic那个点去，函数跑到defer之后的那个点.
/// @brief 4. recover捕获调用函数所有堆栈并非全局捕获，这点我在最开始理解错误
// --------------------------------------------------------------------------

type processor func(data interface{})
func RecoverPanic(fn processor, data interface{}) {
	if err := recover(); err != nil {

		// 打印堆栈日志
		logPanicStack(err)

		// 打印额外日志
		if fn != nil { 
			log.Fatal("====================")
			log.Fatal("宕机额外信息")
			fn(data) 
			log.Fatal("====================")
		}

		// 立刻记录日志
		log.Flush()
		time.Sleep(time.Millisecond)
	}

}


// 非导出，依赖用户日志模块
func logPanicStack(err interface{}) {
	log.Fatal("====================")
	log.Fatal("宕机堆栈")
	log.Fatal("catch panic:'%v'\n",err)
	log.Fatal(string(debug.Stack()))
	log.Fatal("====================")
}


