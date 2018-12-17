/// @file ws_session.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package ws
import	(
	"time"
	"net"
	"gitee.com/jntse/gotoolkit/net/define"
)

// WsConnTask的简单封装，暴露给外部使用
type WsSession struct {
	conn		*WsConnTask 		// 小写不暴露给外部使用
	id 			int
	name 		string
	setUserDefData func(data interface{})
	sendCmd		def.SendHandler
}

func newWsSession(conn *WsConnTask) *WsSession {
	session := &WsSession{conn, conn.id, conn.name, conn.setUserDefdata, conn.sendCmd}
	return session
}

func (w *WsSession) Init() {
}

func (w *WsSession) Id() int {
	return w.id
}

func (w *WsSession) Name() string {
	return w.name
}

func (w *WsSession) SendCmd(msg interface{}) bool {
	return w.sendCmd(msg)
}

func (w *WsSession) VerifySuccess() {
	w.conn.VerifySuccess()
}

func (w *WsSession) IsVerify() bool {
	return w.conn.IsVerify()
}

// 关闭Session(外部调用)
func (w *WsSession) Close() {
	w.conn.quit()
}

func (w *WsSession) DelayClose(delay int64) {
	go func(session* WsSession, delay int64) {
		time.Sleep(time.Second * time.Duration(delay))
		session.Close()
	}(w, delay)
}

func (w *WsSession) SetUserDefData(udata interface{}) {
	w.setUserDefData(udata)
}

func (w *WsSession) UserDefData() interface{} {
	return w.conn.userdata
}

func (w *WsSession) LocalIp() string {
	var addr net.Addr = w.conn.localAddr()
	if addr == nil { return "" }
	return def.NewNetHost(addr.String()).Ip
}

func (w *WsSession) LocalPort() int {
	var addr net.Addr = w.conn.localAddr()
	if addr == nil { return 0 }
	return def.NewNetHost(addr.String()).Port
}

func (w *WsSession) RemoteIp() string {
	var addr net.Addr = w.conn.remoteAddr()
	if addr == nil { return "" }
	return def.NewNetHost(addr.String()).Ip
}

func (w *WsSession) RemotePort() int {
	var addr net.Addr = w.conn.remoteAddr()
	if addr == nil { return 0 }
	return def.NewNetHost(addr.String()).Port
}


