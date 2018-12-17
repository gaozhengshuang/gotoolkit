/// @file ./queue.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-04-23

package eventque
import (
	"time"
	_"sync/atomic"
	"gitee.com/jntse/gotoolkit/log"
)

const (
	KAsynEventQueueStateInvalid = iota + 1
	KAsynEventQueueStateRuning
	KAsynEventQueueStateClosing
	KAsynEventQueueStateShutdown

	//
	kAsynEventQueueDispatch = 10
)

type AsynEventQueue struct {
	ch_recv 	chan IEvent		//
	ch_feedback chan IEvent		//
	ch_capacity	int32
	ch_shutdown	chan int32
	state 		int32
	uid 		int64
}

func NewAsynEventQueue() *AsynEventQueue {
	return &AsynEventQueue{}
}

func (eq *AsynEventQueue) Start(uid int64, cap int32) {
	eq.ch_capacity = cap
	eq.ch_recv = make(chan IEvent, cap)
	eq.ch_feedback = make(chan IEvent, cap)
	eq.ch_shutdown = make(chan int32)
	eq.state = KAsynEventQueueStateRuning
	eq.uid = uid
	go eq.run()
}

func (eq *AsynEventQueue) IsRuning() bool {
	return eq.state == KAsynEventQueueStateRuning
}

// 事件队列满
func (eq *AsynEventQueue) IsFull() bool {
	if eq.state != KAsynEventQueueStateRuning {
		return false
	}
	return eq.EventSize() == eq.ch_capacity
}

// 事件队列大小
func (eq *AsynEventQueue) EventSize() int32 {
	if eq.state != KAsynEventQueueStateRuning {
		return 0
	}
	return int32(len(eq.ch_recv))
}

// 反馈队列大小
func (eq *AsynEventQueue) FeedBackSize() int32 {
	if eq.state != KAsynEventQueueStateRuning {
		return 0
	}
	return int32(len(eq.ch_feedback))
}

// 添加新事件
func (eq *AsynEventQueue) Push(event IEvent) bool {
	if eq.state != KAsynEventQueueStateRuning {
		return false
	}
	//if eq.EventSize() == eq.ch_capacity {
	if eq.EventSize() == eq.ch_capacity {
		log.Error("[AsynEventQueue] uid[%d] EventSize Full [%d]", eq.uid, eq.ch_capacity)
	}
	eq.ch_recv <- event
	return true
	//log.Trace("[AsynEventQueue] uid[%d] Push CurSize[%d]", eq.uid, eq.EventSize())
}

// 关闭队列
func (eq *AsynEventQueue) Shutdown() {
	if eq.state != KAsynEventQueueStateRuning {
		return
	}
	eq.state = KAsynEventQueueStateClosing
	close(eq.ch_shutdown)
	//log.Trace("[AsynEventQueue] uid[%d] Closing Now...", eq.uid)
}

// 队列是否已经关闭了
func (eq *AsynEventQueue) IsShutdown() bool {
	return eq.state == KAsynEventQueueStateShutdown
}

// 反馈队列派发
func (eq *AsynEventQueue) Dispatch() {
	for i := 0; i < kAsynEventQueueDispatch ; i++ {
		if eq.state != KAsynEventQueueStateRuning {
			return
		}
		select {
		case event, open := <-eq.ch_feedback:
			if open == false {
				return
			}
			event.Feedback()
		default:
			return
		}
	}
}

// --------------------------------------------------------------------------
/// @brief 非导出接口
// --------------------------------------------------------------------------
func (eq *AsynEventQueue) run() {
	log.Trace("[AsynEventQueue] uid[%d] start run", eq.uid)
	defer log.Trace("[AsynEventQueue] uid[%d] Quit Done", eq.uid)
	for {
		select {
		case ent, open := <-eq.ch_recv:
			time.Sleep(time.Millisecond)
			if open == false { return }
			if eq.FeedBackSize() == eq.ch_capacity {
				log.Error("[AsynEventQueue] uid[%d] FeedBackSize Full [%d]", eq.uid, eq.ch_capacity)
			}
			ent.Process(eq.ch_feedback)
			break
		case <-eq.ch_shutdown:
			eq.release()
			return
		//default:
		//	time.Sleep(time.Millisecond * 150)	// 空闲时加100毫秒Sleep
		//	if eq.state == KAsynEventQueueStateClosing {
		//		eq.release()
		//		return
		//	}
		//	break
		}
	}
}


// 清理
func (eq *AsynEventQueue) release() {
	//log.Trace("[AsynEventQueue] uid[%d] release ok", eq.uid)
	eq.state = KAsynEventQueueStateShutdown
	close(eq.ch_recv)
	close(eq.ch_feedback)
}


