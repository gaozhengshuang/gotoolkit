/// @file udp_session.go
/// @brief 不完整实现(缺少包编号排序和丢包重发)
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package udp
import (
	"time"
	"net"
	"gitee.com/jntse/gotoolkit/net/define"
)




// UdpConnTask的简单封装，暴露给外部使用
type UdpSession struct {
	conn		*UdpConnTask 		// 小写不暴露给外部使用
	id 			int
	name 		string
	setUserDefData func(data interface{})
	sendCmd		def.SendHandler
}

func newUdpSession(conn *UdpConnTask) *UdpSession {
	session := &UdpSession{conn, conn.id, conn.name, conn.setUserDefdata, conn.sendCmd}
	return session
}

func (u *UdpSession) Init() {
}

func (u *UdpSession) Id() int {
	return u.id
}

func (u *UdpSession) Name() string {
	return u.name
}

func (u *UdpSession) SendCmd(msg interface{}) bool {
	return u.sendCmd(msg)
}

func (u *UdpSession) VerifySuccess() {
	u.conn.VerifySuccess()
}

func (u *UdpSession) IsVerify() bool {
	return u.conn.IsVerify()
}

// 关闭Session(外部调用)
func (u *UdpSession) Close() {
	u.conn.Quit()
}

func (u *UdpSession) DelayClose(delay int64) {
	go func(session* UdpSession, delay int64) {
		time.Sleep(time.Second * time.Duration(delay))
		session.Close()
	}(u, delay)
}

func (u *UdpSession) SetUserDefData(udata interface{}) {
	u.setUserDefData(udata)
}

func (u *UdpSession) UserDefData() interface{} {
	return u.conn.userdata
}

func (u *UdpSession) LocalIp() string {
	var addr net.Addr = u.conn.localAddr()
	if addr == nil { return "" }
	return def.NewNetHost(addr.String()).Ip
}

func (u *UdpSession) LocalPort() int {
	var addr net.Addr = u.conn.localAddr()
	if addr == nil { return 0 }
	return def.NewNetHost(addr.String()).Port
}

func (u *UdpSession) RemoteIp() string {
	var addr net.Addr = u.conn.remoteAddr()
	if addr == nil { return "" }
	return def.NewNetHost(addr.String()).Ip
}

func (u *UdpSession) RemotePort() int {
	var addr net.Addr = u.conn.remoteAddr()
	if addr == nil { return 0 }
	return def.NewNetHost(addr.String()).Port
}


