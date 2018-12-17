/// @file udp_conn.go
/// @brief 不完整实现(缺少包编号排序和丢包重发)
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package udp
import (
	"net"
	"fmt"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/ringbuf"
	"gitee.com/jntse/gotoolkit/net/codec"
	"gitee.com/jntse/gotoolkit/net/define"
	"time"
	"sync"
	"sync/atomic"
	_"encoding/binary"
	"reflect"
)


// --------------------------------------------------------------------------
/// @brief udp连接对象
/// @TODO UdpConnTask golang是强类型语言，protobuf中的int是int32
/// 导致代码中很多地方需要强制类型装换 int->int32, int32->int，有时间将代码中的int使用int32代替
/// @TODO 全部梳理一遍代码，所有的chan都要在一个协程中关闭清理，多线程下极其容易产生隐藏bug
// --------------------------------------------------------------------------
type UdpConnTaskBase struct {
	ip				string
	port			int
	id				int
	name			string
	state			int32
	conn			*net.UDPConn	// net core
	udplisten		*net.UDPConn 	//
	raddr			*net.UDPAddr	// 远端host
	behavior		int32			// accept/connect
}

type UdpConnTaskRWLoop struct {
	//rbuf			[]byte			// read buffer
	rbuf			*ringbuf.Buffer	//
	rbuftmp			[]byte			// TODO: 定义成员变量,避免每次read()构造临时buffer导致频繁GC
	ch_writelen		int32			// 发送队列大小
	ch_write		chan interface{}// TODO: send buffer, send/monitor/外部都会访问(需要加锁)
	ch_quitwloop 	chan int		// 通知退出 write loop
	ch_quitrloop 	chan int		// 通知退出 read loop
	syn_wloop		sync.WaitGroup	// 等待write协程退出
	syn_rloop		sync.WaitGroup	// 等待read协程退出
	ch_recv			chan []byte		// recv buffer
}

type UdpConnTaskExtra struct {
	parser			codec.IBaseParser		// 消息解析器
	legality		def.TcpConnLegality
	netcore			def.INetWork
	session			def.IBaseNetSession		// *TcpSession 暴露外部使用
	cb_event		def.IBaseNetCallback
	ch_eventqueue 	def.EventChan
	dis_eventqueue	bool
	svrchannel		bool			// 服务器task
	userdata		interface{}		// 用户自定义数据
}

type UdpConnTask struct {
	UdpConnTaskBase
	UdpConnTaskRWLoop
	UdpConnTaskExtra
}

// --------------------------------------------------------------------------
/// @brief 导出方法
/// @brief 按照设计UdpConnTask不对外，没有导出接口
// --------------------------------------------------------------------------


// --------------------------------------------------------------------------
/// @brief 非导出方法
// --------------------------------------------------------------------------
func newUdpConnTask(base *UdpConnTaskBase, w_queuelen int32, verify bool, svrchannel bool) (*UdpConnTask) {
	objconn := new(UdpConnTask)
	objconn.UdpConnTaskBase = *base		// TODO: shallow copy(浅拷贝)
	objconn.id = 0
	objconn.state = def.Unconnected

	objconn.ch_writelen = w_queuelen					// TODO: ch_write len 太小很快就会写满, 导致阻塞 
	objconn.ch_write = make(chan interface{}, w_queuelen)
	//objconn.rbuf = make([]byte, 0, def.KCmdRBufMaxSize)
	objconn.rbuf = ringbuf.NewBuffer(def.KCmdRBufMaxSize)
	objconn.rbuftmp = make([]byte, def.KCmdRMaxSize)		// read()需要size不是0 buffer，否则返回0
	objconn.ch_quitwloop = make(chan int, 1)
	objconn.ch_quitrloop = make(chan int, 1)
	objconn.legality.Init()
	if verify == true {
		objconn.legality.TimeOut = time.Now().Unix() + def.KLegalityVerifyTimeOut
		objconn.legality.VerifyState = def.ConnVerifying
	}
	objconn.session = newUdpSession(objconn)
	objconn.svrchannel = svrchannel
	objconn.raddr = nil
	objconn.ch_recv = make(chan []byte, def.KServerRecvQueueSize)

	return objconn
}

func (u *UdpConnTask) Init(parser codec.IBaseParser, netcore def.INetWork, dis_eventqueue bool) {
	u.parser = parser
	u.netcore = netcore
	u.id = int(netcore.GenerateTaskId())
	u.cb_event = netcore.EventBaseCb()
	u.ch_eventqueue = netcore.EventQueue()
	u.dis_eventqueue = dis_eventqueue
}

// 绑定用户数据，以函数指针方式暴露接口暴露
func (u *UdpConnTask) setUserDefdata(udata interface{})	{
	u.userdata = udata
}


// LocalAddr 获取本地地址
func (u *UdpConnTask) localAddr() net.Addr {
	if u.conn != nil { return (u.conn).LocalAddr() }
	return nil
}


// RemoteAddr 获取远端地址
func (u *UdpConnTask) remoteAddr() net.Addr {
	if u.conn != nil { return (u.conn).RemoteAddr() }
	return nil
}


// 关闭退出 TODO: 有可能此时read和write报错通知MonitorLoop开始走清理，导致chan monitorloop被清理掉
// 解决方案: 不使用 chan monitorloop，改为使用Closing状态来结束task
func (u *UdpConnTask) Quit() {
	stat := atomic.LoadInt32(&u.state)
	if stat == def.Closed || stat == def.Closing {
		log.Error("sid[%d] is 'Closed' or 'Closing' stat[%d]", u.id, stat)
		return
	}

	atomic.StoreInt32(&u.state, def.Closing)
}

// 最后调用，释放所有协程和资源
func (u *UdpConnTask) cleanup() {
	u.quitwriteLoop()	// 退出wloop
	u.waitWloopQuit()	// 等待wloop
	u.quitreadLoop()		// 退出rloop
	u.waitRloopQuit()	// 等待rloop
	u.netcore.OnSessionClose(u)

	//
	atomic.StoreInt32(&u.state, def.Closed)
	u.conn = nil
	u.rbuf = nil
	u.rbuftmp = nil
	u.session = nil
	u.udplisten = nil

	close(u.ch_write)
	u.ch_write = nil
	close(u.ch_quitwloop)
	u.ch_quitwloop = nil
	close(u.ch_quitrloop)
	u.ch_quitrloop = nil
	close(u.ch_recv)
	u.ch_recv = nil

}

func (u *UdpConnTask) isNilSession() bool {
	if u.session == nil || reflect.ValueOf(u.session).IsNil() {
		return true
	}
	return false
}


// 连接
func (u *UdpConnTask) connect() bool {
	host := fmt.Sprintf("%s:%d",u.ip, u.port)
	//*/
	udpAddr, err_resolve := net.ResolveUDPAddr("udp", host)
	if err_resolve != nil {
		log.Error("'%s' ResolveUDPAddr error:%v", u.name, err_resolve)
		return false
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)	// UDPConn, error
	if err != nil {
		log.Error("sid[%d] can't connect to [%s] err=%s", u.id, host, err.Error())
		return false
	}
	// UDPConn专属设置接口
	//conn.SetKeepAlive(true)
	//conn.SetKeepAlivePeriod(def.KTcpKeepAlivePeriod)
	//conn.SetNoDelay(true)
	conn.SetWriteBuffer(def.KWriteBufferSize)
	conn.SetReadBuffer(def.KReadBufferSize)

	u.conn = conn
	atomic.StoreInt32(&u.state, def.Connected)
	u.syn_wloop.Add(1)
	u.syn_rloop.Add(1)
	u.startWritereadLoop()
	u.session = newUdpSession(u)
	return true
}

// Accept一个连接
func (u *UdpConnTask) accpeted() {
	atomic.StoreInt32(&u.state, def.Connected)
	u.syn_wloop.Add(1)
	u.syn_rloop.Add(1)
	u.startWritereadLoop()
	u.session = newUdpSession(u)
}


// TODO: 	使用通道来作为发送缓冲区，和传统缓冲区有差别
// 			0.外部协程访问
// 			1.以包作为单位，而非连续内存
//			2.致命问题: 如果ch_write通道满了，会阻塞发送操作所在协程，例如主线程
//			3.一般来说一个玩家发送队列不会积累太多消息包，发生了2情况说明这个玩家网络已经延迟很长了
//			4.致命问题: 例如发送一个1024byte包没有一次性发送完毕，不支持把剩下未发送的内容加入到发送队列最前面
//			5.最终结论: 添加一个传统的发送缓冲区
func (u *UdpConnTask) sendCmd(msg interface{}) bool {
	if atomic.LoadInt32(&u.state) != def.Connected {
		log.Error("sid[%d] conn is Closed send fail", u.id)
		return false
	}

	msgnum := int32(len(u.ch_write))
	if msgnum >= u.ch_writelen {
		u.Quit()
		panic(fmt.Sprintf("sid:'%d' ch_write is full, force close!!!", u.id))
	}

	u.ch_write <- msg
	return true
}


// 重置状态准备重连
// TODO: 不一定需要重新make变量，频繁断开可能会导致内存爆炸,频繁GC等问题
func (u *UdpConnTask) reset() {
	u.conn = nil
	atomic.StoreInt32(&u.state, def.Unconnected)

	u.ch_write = make(chan interface{}, u.ch_writelen)
	//u.rbuf = make([]byte, 0, def.KCmdRBufMaxSize)
	if u.rbuf == nil { u.rbuf = ringbuf.NewBuffer(def.KCmdRBufMaxSize) }
	u.rbuf.Reset()
	u.rbuftmp = make([]byte, def.KCmdRMaxSize)	// read()需要size不是0 buffer
	u.ch_quitwloop = make(chan int, 1)
	u.ch_quitrloop = make(chan int, 1)
}


func (u *UdpConnTask) waitRloopQuit() {
	u.syn_rloop.Wait()
}

func (u *UdpConnTask) waitWloopQuit() {
	u.syn_wloop.Wait()
}

// 启动收发Loop
func (u *UdpConnTask) startWritereadLoop() {
	go u.writeLoop()
	go u.readLoop()
}


func (u* UdpConnTask) quitreadLoop() {
	if u.conn != nil { u.conn.Close(); }	// 使用阻塞read需要Close()来触发read内部报错解除阻塞状态，退出readloop
	u.ch_quitrloop <- 1
}


func (u* UdpConnTask) quitwriteLoop() {
	u.ch_quitwloop <- 1
}


func (u *UdpConnTask) writeLoop()	{
	defer u.syn_wloop.Done()
	//defer log.Info("sid[%d] writeLoop Quit", u.id)
	//log.Info("enter writeLoop[%d]", util.GetRoutineID())

	// TODO: 这里不使用原子读应该也ok，担心影响效率
	for u.state == def.Connected {
		if u.svrchannel == true {
			time.Sleep(time.Microsecond*1)
		}else {
			time.Sleep(time.Millisecond*10)
		}
		select {
		case <-u.ch_quitwloop:
			return
		case msg, open := <-u.ch_write:
			if open == false {
				log.Info("sid[%d] ch_write has been Closed", u.id)
				return
			}

			data , ok := u.parser.PackMsg(msg)
			if ok == false {
				log.Fatal("ProtoMsg PackMsg Fail")
				return
			}

			//u.wbuf = append(u.wbuf, data...)
			_, err := u.write(data)
			if err != nil {
				log.Error("sid[%d] WriteError:'%s'", u.id, err)
				return
			}
		}
	}
}

func (u* UdpConnTask) readLoop() {
	defer u.syn_rloop.Done()
	//defer log.Info("sid[%d] readLoop Quit", u.id)
	//log.Info("enter readLoop[%d]", util.GetRoutineID())

	// TODO: 这里不使用原子读应该也ok，担心影响效率
	for u.state == def.Connected {
		if u.svrchannel == true {
			time.Sleep(time.Microsecond*1)
		}else {
			time.Sleep(time.Millisecond*10)
		}
		select {
		case <-u.ch_quitrloop:
			return
		default:
			_, err := u.read()
			if err != nil	{
				log.Error("sid[%d] readError:%s", u.id, err)
				return
			}
		}
	}
}

func (u *UdpConnTask) read()  (bool, error) {
	for i:=0; i < 10; i++ {
		//u.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))	// 设置读超时, nonblock read
		len := 0
		var err error = nil
		if u.behavior == def.Acceptor {
			len, err = u.readFromListener(u.rbuftmp)		// read()需要size不是0的buffer, 否则永远返回0
		}else {
			len, err = u.conn.Read(u.rbuftmp)		// read()需要size不是0的buffer, 否则永远返回0
		}
		if err != nil {
			if nerr, convertok := err.(net.Error); convertok && nerr.Timeout() { 
				return true , nil
			}
			u.Quit()
			return false, err
		}

		// TODO: rbuf无限增大，有风险导致内存耗尽
		//u.rbuf = append(u.rbuf, u.rbuftmp[:len]...)
		u.rbuf.Write(u.rbuftmp[:len])

		// all unpack done
		for ;; {
			msg, msghandler, err := u.parser.UnPackMsg(u.rbuf) 
			if err != nil {
				u.Quit()
				break
			}

			if msg == nil || msghandler == nil {
				break
			}

			// 成功解析协议信任该连接
			u.VerifySuccess()

			//TODO: 1. 直接回调--每个客户端消息处理都在单独的协程中 2. 事件通知--发送到主逻辑协程处理
			if true == u.dis_eventqueue {
				msghandler.Handler(u.session , msg)
			}else {
				msgEvent := &def.MsgDispatchEvent{Session:u.session, Msg:msg, Handler:msghandler.Handler}
				u.ch_eventqueue <- msgEvent
			}
		}
	}
	return true, nil
}

func (u *UdpConnTask) write(data []byte) (bool, error) {
	len_total, len_write := len(data), 0
	for i:=0; i < 10; i++ {
		//u.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))	// 设置写超时, Nonblock write
		nleft := len(data)
		wlen := 0
		var err error = nil
		if u.behavior  == def.Acceptor {
			wlen, err = (u.udplisten).WriteToUDP(data, u.raddr)
		}else {
			wlen, err = (u.conn).Write(data)
		}

		len_write += wlen
		if err != nil {
			u.Quit()
			return false, err
		}
		if wlen < nleft { 
			data = data[wlen:nleft]
			continue
		}
		break
	}

	if len_write != len_total {
		errlog := fmt.Sprintf("sid [%d] [%s] write not send all data !!! want:%d real:%d", u.id, u.name, len_total, len_write)
		log.Error(errlog)
		//panic(errlog)
	}
	//log.Info("sid[%d] real send data len=%d", u.id, len_write)
	return true, nil
}

// 
func (u *UdpConnTask) setRemoteAddr(addr *net.UDPAddr) {
	u.raddr = addr
}

// 从udp listener 中读取数据，采用阻塞模式
func (u *UdpConnTask) readFromListener(b []byte) (int, error) {
	select {
	case data, open := <-u.ch_recv:
		if open == false { 
			return 0, fmt.Errorf("read from udp chan 'ch_recv' is Closed")
		}
		copy(b, data)
		return len(data), nil
	default:
		break
	}
	return 0, nil
	//data, open := <-u.ch_recv
	//if open == false { 
	//	return 0, fmt.Errorf("read from udp chan 'ch_recv' is Closed")
	//}
	//copy(b, data)
	//return len(data), nil
}

func (u *UdpConnTask) pushRecvData(b []byte) {
	tmp := make([]byte, len(b))	// TODO: chan传入的是切片引用，这里需要重新new一份切片
	copy(tmp, b)
	u.ch_recv <- tmp
}


// --------------------------------------------------------------------------
/// @brief 导出接口
// --------------------------------------------------------------------------
// 验证通过
func (u *UdpConnTask) VerifySuccess() {
	if atomic.LoadInt32(&u.legality.VerifyState) == def.ConnVerifying {
		atomic.StoreInt32(&u.legality.VerifyState, def.ConnVerifySuccess)
		log.Info("sid[%d] conn verify ok", u.id)
	}
}

func (u *UdpConnTask) IsVerify() bool {
	return atomic.LoadInt32(&u.legality.VerifyState) == def.ConnVerifySuccess
}

func (u *UdpConnTask) GetSession() def.IBaseNetSession {
	return u.session
}


// --------------------------------------------------------------------------
/// @brief OnConnect和OnClose有2中方式回调 
//	1.在MonitorLoop协程直接回调(多线程)
//	2.通过事件发送到主逻辑协程处理
// --------------------------------------------------------------------------
func (u *UdpConnTask) OnClose() {
	if u.isNilSession() { 
		return
	}

	// 连接成功才需要回调OnClose接口
	if true ==  u.dis_eventqueue {
		u.cb_event.OnClose(u.session)
	}else {
		msgEvent := &def.NetCloseEvent{Session:u.session, Handler:u.cb_event.OnClose}
		u.ch_eventqueue <- msgEvent
	}
}

func (u *UdpConnTask) OnConnect() {
	if u.isNilSession() {
		return
	}
	
	if true == u.dis_eventqueue {
		u.cb_event.OnConnect(u.session)
	}else {
		msgEvent := &def.NetConnectEvent{Session:u.session, Handler:u.cb_event.OnConnect}
		u.ch_eventqueue <- msgEvent
	}
}


