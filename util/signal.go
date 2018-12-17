/// @file signal.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-01

package util
import "fmt"
import "os"
import "os/signal"
//import "syscall"
import "sync"
import "time"


type SignalHandle func(signal os.Signal)
type Signal struct {
	handlers    map[os.Signal]SignalHandle
	ch          chan os.Signal
	synquit     sync.WaitGroup
	running     bool
}


func CreateNewSignal() *Signal {
	t := new(Signal)
	return t
}


func (t *Signal) Init() {
	t.handlers = make(map[os.Signal]SignalHandle)
	t.ch = make(chan os.Signal, 1)
}


// --------------------------------------------------------------------------
/// @brief 
// syscall.SIGINT
// syscall.SIGQUIT
// syscall.SIGABRT
// syscall.SIGTERM
// syscall.SIGPIPE
// syscall.SIGHUP
// --------------------------------------------------------------------------
func (t* Signal) Regist(id os.Signal, handler SignalHandle ) {
	t.handlers[id] = handler
	signal.Notify(t.ch, id)
}

func (t* Signal) Start() {
	if t.running { panic("Signal is running already") }
	t.running = true
	t.synquit.Add(1)
	go t.run()
}


func (t *Signal) run() {
	defer t.synquit.Done()
	for ;; {
		time.Sleep(time.Millisecond * 10)
		select {
		case id, open := <-t.ch:
			if open == false {
				fmt.Println("util.Signal coroutine quit")
				return
			}
			t.handle(id)
		}
	}
}


func (t *Signal) handle(id os.Signal) {
	fmt.Printf("\nreceive a signal '%s'\n", id)
	f, rfind := t.handlers[id]
	if rfind == true { f(id) }
	//switch signal {
	//case syscall.SIGHUP:
	//case syscall.SIGINT:
	//case syscall.SIGQUIT:
	//}
}


// 完全退出
func (t *Signal) Close() {
	signal.Stop(t.ch)	// 停止接收信号
	close(t.ch)
	t.synquit.Wait()
	t.running = false
}


