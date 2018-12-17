/// @file event.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package def

type IBaseNetCallback interface {
	OnClose(session IBaseNetSession)
	OnConnect(session IBaseNetSession)
}


type EventChan = chan ISessionEvent
type ISessionEvent interface {
	InitEvent()
	Process()
}


// --------------------------------------------------------------------------
/// @brief 网络 OnConnect 事件
/// @brief Session用接口虽然很通用但是类型断言有性能损失，目前暂时使用具体类型吧
// --------------------------------------------------------------------------
type NetConnectEvent struct {
	Session		IBaseNetSession
	Handler 	func(session IBaseNetSession)
}
func (ev *NetConnectEvent) InitEvent() {
}

func (ev *NetConnectEvent) Process() {
	ev.Handler(ev.Session)
}


// --------------------------------------------------------------------------
/// @brief 网络 OnClose 事件
/// @brief 
// --------------------------------------------------------------------------
type NetCloseEvent struct {
	Session		IBaseNetSession
	Handler     func(session IBaseNetSession)
}
func (ev *NetCloseEvent) InitEvent() {
}

func (ev *NetCloseEvent) Process() {
	ev.Handler(ev.Session)
}


// --------------------------------------------------------------------------
/// @brief 网络 OnError 事件
/// @brief 
// --------------------------------------------------------------------------
type NetErrorEvent struct {
	Session		IBaseNetSession
	Handler     func(session IBaseNetSession)
}
func (ev *NetErrorEvent) InitEvent() {
}

func (ev *NetErrorEvent) Process() {
	ev.Handler(ev.Session)
}

// --------------------------------------------------------------------------
/// @brief 消息派发处理事件
/// @brief Session用接口虽然很通用但是类型断言有性能损失，目前暂时使用具体类型吧
// --------------------------------------------------------------------------
type MsgDispatchEvent struct {
	Session 	IBaseNetSession
	Msg  		interface{}
	Handler 	MsgHandle
}
func (ev *MsgDispatchEvent) InitEvent() {
}

func (ev *MsgDispatchEvent) Process() {
	ev.Handler(ev.Session, ev.Msg)
}


// --------------------------------------------------------------------------
/// @brief http 应答事件
// --------------------------------------------------------------------------
type HttpResponseEvent struct {
	Msg interface{}
	Handler func(msg interface{})
}

func (ev *HttpResponseEvent) InitEvent() {
}

func (ev *HttpResponseEvent) Process() {
	ev.Handler(ev.Msg)
}

