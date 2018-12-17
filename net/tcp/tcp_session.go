/// @file tcp_session.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package tcp
import (
	"time"
	"net"
	"gitee.com/jntse/gotoolkit/net/define"
)

// --------------------------------------------------------------------------
/// @brief TcpConnTask的简单封装，暴露给外部使用
/// @brief 所有成员非导出，使用接口访问
// --------------------------------------------------------------------------
type TcpSession struct {
	conn		*TcpConnTask
	id 			int
	name 		string
	setUserDefData func(data interface{})
	sendCmd		def.SendHandler
}

func newTcpSession(conn *TcpConnTask) *TcpSession {
	session := &TcpSession{conn, conn.id, conn.name, conn.setUserDefdata, conn.sendCmd}
	return session
}

func (t *TcpSession) Init() {
}

func (t *TcpSession) Id() int {
	return t.id
}

func (t *TcpSession) Name() string {
	return t.name
}

func (t *TcpSession) SendCmd(msg interface{}) bool {
	return t.sendCmd(msg)
}

func (t *TcpSession) VerifySuccess() {
	t.conn.VerifySuccess()
}

func (t *TcpSession) IsVerify() bool {
	return t.conn.IsVerify()
}

// 关闭Session(外部调用)
func (t *TcpSession) Close() {
	t.conn.quit()
}

func (t *TcpSession) DelayClose(delay int64) {
	go func(session* TcpSession, delay int64) {
		time.Sleep(time.Second * time.Duration(delay))
		session.Close()
	}(t, delay)
}

func (t *TcpSession) SetUserDefData(udata interface{}) {
	t.setUserDefData(udata)
}

func (t *TcpSession) UserDefData() interface{} {
	return t.conn.userdata
}

func (t *TcpSession) LocalIp() string {
	var addr net.Addr = t.conn.localAddr()
	if addr == nil { return "" }
	return def.NewNetHost(addr.String()).Ip
}

func (t *TcpSession) LocalPort() int {
	var addr net.Addr = t.conn.localAddr()
	if addr == nil { return 0 }
	return def.NewNetHost(addr.String()).Port
}

func (t *TcpSession) RemoteIp() string {
	var addr net.Addr = t.conn.remoteAddr()
	if addr == nil { return "" }
	return def.NewNetHost(addr.String()).Ip
}

func (t *TcpSession) RemotePort() int {
	var addr net.Addr = t.conn.remoteAddr()
	if addr == nil { return 0 }
	return def.NewNetHost(addr.String()).Port
}


