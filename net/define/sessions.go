/// @file sessions.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package def
import (
	"sync"
	"gitee.com/jntse/gotoolkit/log"
	_"fmt"
	"reflect"
	"math"
	"sync/atomic"
)

// --------------------------------------------------------------------------
/// @brief generate TcpConnTask Id
// --------------------------------------------------------------------------
//var g_ConnTaskId int32 = 0
//func genConnTaskId() int32 {
//	return atomic.AddInt32(&g_ConnTaskId, 1)
//}

type NetSessionPool struct {
	spools map[int]IBaseNetSession
	slocker *sync.Mutex		// *sync.RWMutex
	taskid int32


	rlocker *sync.Mutex	// *sync.RWMutex
	recycleid []int32
}


func NewNetSessionPool() *NetSessionPool {
	pool := &NetSessionPool{ spools:make(map[int]IBaseNetSession), slocker:&sync.Mutex{}, taskid:0 }
	return pool
}

func (sp *NetSessionPool) Init() {
	sp.spools = make(map[int]IBaseNetSession)
	sp.slocker = &sync.Mutex{}
	sp.taskid = 0

	sp.rlocker = &sync.Mutex{}
	sp.recycleid = make([]int32, 0, 1000000)
}

func (sp *NetSessionPool) Size() int32 {
	return int32(len(sp.spools))
}

func (sp *NetSessionPool) GenerateTaskId() int32 {	
	//return util.UUID()
	if atomic.LoadInt32(&sp.taskid) < math.MaxInt32 {
		id := atomic.AddInt32(&sp.taskid, 1)
		return id
	}
	
	//
	id := int32(-1)
	sp.rlocker.Lock()
	if len(sp.recycleid) > 0 { 
		id = sp.recycleid[0]
		sp.recycleid = sp.recycleid[1:]
	}
	sp.rlocker.Unlock()
	return id
}

func (sp *NetSessionPool) AddSession(s IBaseNetSession) bool {
	if s == nil || reflect.ValueOf(s).IsNil() { return false }
	sp.slocker.Lock()
	if _, ok := sp.spools[s.Id()]; ok == true {
		sp.slocker.Unlock()
		return false
	}
	sp.spools[s.Id()] = s
	sp.slocker.Unlock()
	return true
}

func (sp *NetSessionPool) DelSession(s IBaseNetSession) {
	if s == nil || reflect.ValueOf(s).IsNil() { return }
	sp.slocker.Lock()
	delete(sp.spools, s.Id())
	sp.slocker.Unlock()

	//
	sp.rlocker.Lock()
	id := int32(s.Id())
	sp.recycleid = append(sp.recycleid, id)
	sp.rlocker.Unlock()
}


func (sp *NetSessionPool) FindSession(sid int) IBaseNetSession {
	sp.slocker.Lock()
	s, ok := sp.spools[sid]
	sp.slocker.Unlock()
	if ok == true {	return s }
	return nil
}


func (sp *NetSessionPool) SendMsg(sid int, msg interface{}) bool {
	sp.slocker.Lock()
	s, ok := sp.spools[sid]
	if ok == false {
		sp.slocker.Unlock()
		return false
	}
	sp.slocker.Unlock()
	return s.SendCmd(msg)
}

// 新会话建立
func (sp *NetSessionPool) OnSessionEstablished(conn IBaseConnTask) {
	sp.AddSession(conn.GetSession())
	conn.OnConnect()
	log.Trace("当前会话数:%d", sp.Size())
}

// 会话断开
func (sp *NetSessionPool) OnSessionClose(conn IBaseConnTask) {
	sp.DelSession(conn.GetSession())
	conn.OnClose()
	log.Trace("当前会话数:%d", sp.Size())
}


