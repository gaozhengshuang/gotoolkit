/// @file tcp_conn.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package tcp
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
/// @brief tcp连接对象
/// @TODO TcpConnTask golang是强类型语言，protobuf中的int是int32
/// 导致代码中很多地方需要强制类型装换 int->int32, int32->int，有时间将代码中的int使用int32代替
/// @TODO 全部梳理一遍代码，所有的chan都要在一个协程中关闭清理，多线程下极其容易产生隐藏bug
// --------------------------------------------------------------------------

// 基础属性
type TcpConnTaskBase struct {
	ip				string
	port			int
	id				int
	name			string
	state			int32
	conn			*net.TCPConn	// net core
	behavior		int32			// accept/connect
}

// 读写Loop
type TcpConnTaskLoop struct {
	//rbuf			[]byte			// read buffer
	rbuf			*ringbuf.Buffer	// read buffer 环形缓冲区
	rbuftmp			[]byte			// TODO: 定义成员变量,避免每次read()构造临时buffer导致频繁GC
	ch_writelen		int32			// 发送队列大小
	ch_write		chan interface{}// TODO: send buffer, send/monitor/外部都会访问(需要加锁)
	//wbufleft		*ringbuf.Buffer	// 不完全write剩余数据
	//wbuf			[]byte			// 发送缓冲区
	//wbuftmp		[]byte			// 临时发送缓冲区，避免频繁GC
	ch_quitwloop 	chan int		// 通知退出 write loop
	ch_quitrloop 	chan int		// 通知退出 read loop
	syn_wloop		sync.WaitGroup	// 等待write协程退出
	syn_rloop		sync.WaitGroup	// 等待read协程退出
}

// 额外属性
type TcpConnTaskExtra struct {
	parser			codec.IBaseParser	// 消息解析器
	legality		def.TcpConnLegality
	session			def.IBaseNetSession	// *TcpSession 暴露外部使用
	netcore			def.INetWork
	cb_event		def.IBaseNetCallback
	ch_eventqueue 	def.EventChan
	dis_eventqueue	bool
	svrchannel		bool				// 服务器task
	userdata		interface{}			// 用户自定义数据
}

type TcpConnTask struct {
	TcpConnTaskBase
	TcpConnTaskLoop
	TcpConnTaskExtra
}

// --------------------------------------------------------------------------
/// @brief 非导出方法
// --------------------------------------------------------------------------

// 新建实例
func newTcpConnTask(base *TcpConnTaskBase, w_queuelen int32, verify bool, svrchannel bool) (*TcpConnTask) {
	objconn := new(TcpConnTask)
	objconn.TcpConnTaskBase = *base		// TODO: shallow copy(浅拷贝)
	objconn.id = 0
	objconn.state = def.Unconnected

	objconn.ch_writelen = w_queuelen					// TODO: ch_write len 太小很快就会写满, 导致阻塞 
	objconn.ch_write = make(chan interface{}, w_queuelen)
	//objconn.rbuf = make([]byte, 0, def.KCmdRBufMaxSize)
	objconn.rbuf = ringbuf.NewBuffer(def.KCmdRBufMaxSize)
	objconn.rbuftmp = make([]byte, def.KCmdRDMaxSize)		// read()需要size不是0 buffer，否则返回0
	//objconn.wbuf = make([]byte, 0, KCmdWBufMaxSize)
	//objconn.wbuftmp = make([]byte, KCmdWRMaxSize)		// 
	//objconn.wbufleft = ringbuf.NewBuffer(def.KCmdWRMaxSize)
	objconn.ch_quitwloop = make(chan int, 1)
	objconn.ch_quitrloop = make(chan int, 1)
	objconn.legality.Init()
	if verify == true {
		objconn.legality.TimeOut = time.Now().Unix() + def.KLegalityVerifyTimeOut
		objconn.legality.VerifyState = def.ConnVerifying
	}
	objconn.session = newTcpSession(objconn)
	objconn.svrchannel = svrchannel
	//log.Trace("================ w_queuelen:%d svrchannel:%t ================",w_queuelen,svrchannel)
	return objconn
}

func (t *TcpConnTask) Init(parser codec.IBaseParser, netcore def.INetWork, dis_eventqueue bool) {
	t.parser = parser
	t.netcore = netcore
	t.id = int(netcore.GenerateTaskId())
	t.cb_event = netcore.EventBaseCb()
	t.ch_eventqueue = netcore.EventQueue()
	t.dis_eventqueue = dis_eventqueue
}

// 绑定用户数据，以函数指针方式暴露接口暴露
func (t *TcpConnTask) setUserDefdata(udata interface{})	{
	t.userdata = udata
}


// LocalAddr 获取本地地址
func (t *TcpConnTask) localAddr() net.Addr {
	if t.conn != nil { return (t.conn).LocalAddr() }
	return nil
}


// RemoteAddr 获取远端地址
func (t *TcpConnTask) remoteAddr() net.Addr {
	if t.conn != nil { return (t.conn).RemoteAddr() }
	return nil
}


// 关闭退出 TODO: 有可能此时read和Write报错通知MonitorLoop开始走清理，导致chan monitorloop被清理掉
// 解决方案: 不使用 chan monitorloop，改为使用Closing状态来结束task
func (t *TcpConnTask) quit() {
	stat := atomic.LoadInt32(&t.state)
	if stat == def.Closed || stat == def.Closing {
		log.Error("sid[%d] is 'Closed' or 'Closing' stat[%d]", t.id, stat)
		return
	}

	atomic.StoreInt32(&t.state, def.Closing)
}

// 最后调用，释放所有协程和资源
func (t *TcpConnTask) cleanup() {
	t.quitwriteLoop()	// 退出wloop
	t.waitWloopQuit()	// 等待wloop
	t.quitreadLoop()		// 退出rloop
	t.waitRloopQuit()	// 等待rloop
	//t.OnClose()
	t.netcore.OnSessionClose(t)

	//
	atomic.StoreInt32(&t.state, def.Closed)
	//if t.conn != nil { t.conn.Close() }
	t.conn = nil
	t.rbuf = nil
	t.rbuftmp = nil
	//t.wbuf = nil
	//t.wbuftmp = nil
	//t.wbufleft = nil
	t.session = nil

	close(t.ch_write)
	t.ch_write = nil
	close(t.ch_quitwloop)
	t.ch_quitwloop = nil
	close(t.ch_quitrloop)
	t.ch_quitrloop = nil

}

// 连接
func (t *TcpConnTask) connect() bool {
	host := fmt.Sprintf("%s:%d",t.ip, t.port)
	//*/
	tcpAddr, err_resolve := net.ResolveTCPAddr("tcp", host)
	if err_resolve != nil {
		log.Error("'%s' ResolveTCPAddr error:%v", t.name, err_resolve)
		return false
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)	// TCPConn, error
	if err != nil {
		log.Error("sid[%d] can't connect to [%s] err=%s", t.id, host, err.Error())
		return false
	}
	// TCPConn专属设置接口
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(def.KTcpKeepAlivePeriod)
	conn.SetNoDelay(true)
	conn.SetWriteBuffer(def.KWriteBufferSize)
	conn.SetReadBuffer(def.KReadBufferSize)
	/*/
	conn, err := net.Dial("tcp", host)
	if err != nil {
		log.Error("sid[%d] can't connect to [%s] err=%s", t.id, host, err.Error())
		return false
	}
	//*/

	t.conn = conn
	atomic.StoreInt32(&t.state, def.Connected)
	t.syn_wloop.Add(1)
	t.syn_rloop.Add(1)
	t.startWritereadLoop()
	t.session = newTcpSession(t)
	return true
}

// Accept一个连接
func (t *TcpConnTask) accpeted() {
	atomic.StoreInt32(&t.state, def.Connected)
	t.syn_wloop.Add(1)
	t.syn_rloop.Add(1)
	t.startWritereadLoop()
	t.session = newTcpSession(t)
}

// TODO: 	使用通道来作为发送缓冲区，和传统缓冲区有差别
// 			0.外部协程访问
// 			1.以包作为单位，而非连续内存
//			2.致命问题: 如果ch_write通道满了，会阻塞发送操作所在协程，例如主线程
//			3.一般来说一个玩家发送队列不会积累太多消息包，发生了2情况说明这个玩家网络已经延迟很长了
//			4.致命问题: 例如发送一个1024byte包没有一次性发送完毕，不支持把剩下未发送的内容加入到发送队列最前面
//			5.最终结论: 添加一个传统的发送缓冲区
func (t *TcpConnTask) sendCmd(msg interface{}) bool {
	if atomic.LoadInt32(&t.state) != def.Connected {
		log.Error("sid[%d] conn is Closed send fail", t.id)
		return false
	}

	msgnum := int32(len(t.ch_write))
	if msgnum >= t.ch_writelen {
		t.quit()
		panic(fmt.Sprintf("sid:'%d' ch_write is full, force close!!!", t.id))
	}

	t.ch_write <- msg
	return true
}

// 重置状态准备重连
// TODO: 不一定需要重新make变量，频繁断开可能会导致内存爆炸,频繁GC等问题
func (t *TcpConnTask) reset() {
	t.conn = nil
	atomic.StoreInt32(&t.state, def.Unconnected)

	t.ch_write = make(chan interface{}, t.ch_writelen)
	//t.rbuf = make([]byte, 0, def.KCmdRBufMaxSize)
	if t.rbuf == nil { t.rbuf = ringbuf.NewBuffer(def.KCmdRBufMaxSize) }
	t.rbuf.Reset()
	t.rbuftmp = make([]byte, def.KCmdRDMaxSize)	// read()需要size不是0 buffer
	//if t.wbufleft == nil { t.wbufleft = ringbuf.NewBuffer(def.KCmdWRMaxSize) }
	//t.wbufleft.Reset();
	//t.wbuf = make([]byte, 0, KCmdWBufMaxSize)
	//t.wbuftmp = make([]byte, KCmdWRMaxSize)	// read()需要size不是0 buffer
	t.ch_quitwloop = make(chan int, 1)
	t.ch_quitrloop = make(chan int, 1)
}


func (t *TcpConnTask) waitRloopQuit() {
	t.syn_rloop.Wait()
}

func (t *TcpConnTask) waitWloopQuit() {
	t.syn_wloop.Wait()
}

// 启动收发Loop
func (t *TcpConnTask) startWritereadLoop() {
	go t.writeLoop()
	go t.readLoop()
}


func (t* TcpConnTask) quitreadLoop() {
	if t.conn != nil { t.conn.Close(); }	// 使用阻塞read需要Close()来触发read内部报错解除阻塞状态，退出readloop
	t.ch_quitrloop <- 1
}


func (t* TcpConnTask) quitwriteLoop() {
	t.ch_quitwloop <- 1
}


func (t *TcpConnTask) writeLoop()	{
	defer t.syn_wloop.Done()
	//defer log.Info("sid[%d] writeLoop Quit", t.id)
	//log.Info("enter writeLoop[%d]", util.GetRoutineID())

	// TODO: 这里不使用原子读应该也ok，担心影响效率
	for t.state == def.Connected {
		if t.svrchannel == true {
			time.Sleep(time.Microsecond*1)
		}else {
			time.Sleep(time.Millisecond*10)
		}

		//if t.wbufleft.Len() != 0 {
		//	select {
		//	case <-t.ch_quitwloop:
		//		return
		//	default:
		//		data := t.wbufleft.ReadByReference(t.wbufleft.Len());
		//		_, err := t.write(data)
		//		if err != nil {
		//			log.Error("sid[%d] WriteError:'%s'", t.id, err)
		//			return
		//		}
		//		continue;
		//	}
		//}

		select {
		case <-t.ch_quitwloop:
			return
		case msg, open := <-t.ch_write:
			if open == false {
				log.Info("sid[%d] ch_write has been Closed", t.id)
				return
			}

			data , ok := t.parser.PackMsg(msg)
			if ok == false {
				log.Fatal("ProtoMsg PackMsg Fail")
				return
			}

			//t.wbuf = append(t.wbuf, data...)
			_, err := t.write(data)
			if err != nil {
				log.Error("sid[%d] WriteError:'%s'", t.id, err)
				return
			}
		}
	}
}

func (t* TcpConnTask) readLoop() {
	defer t.syn_rloop.Done()
	//defer log.Info("sid[%d] readLoop Quit", t.id)
	//log.Info("enter readLoop[%d]", util.GetRoutineID())

	// TODO: 这里不使用原子读应该也ok，担心影响效率
	for t.state == def.Connected {
		if t.svrchannel == true {
			time.Sleep(time.Microsecond*1)
		}else {
			time.Sleep(time.Millisecond*10)
		}
		select {
		case <-t.ch_quitrloop:
			return
		default:
			_, err := t.read()
			if err != nil	{
				log.Error("sid[%d] readError:%s", t.id, err)
				return
			}
		}
	}
}

func (t *TcpConnTask) read()  (bool, error) {
	for i:=0; i < 10; i++ {
		//t.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))	// 设置读超时, nonblock read
		len, err := t.conn.Read(t.rbuftmp)		// read()需要size不是0的buffer, 否则永远返回0
		if err != nil || len == 0 {
			if nerr, convertok := err.(net.Error); convertok && nerr.Timeout() { 
				return true , nil
			}
			t.quit()
			return false, err
		}

		// TODO: rbuf无限增大，有风险导致内存耗尽
		//t.rbuf = append(t.rbuf, t.rbuftmp[:len]...)
		t.rbuf.Write(t.rbuftmp[:len])

		// all unpack done
		for ;; {
			msg, msghandler, err := t.parser.UnPackMsg(t.rbuf) 
			if err != nil {
				t.quit()
				break
			}

			if msg == nil || msghandler == nil {
				break
			}

			// 成功解析协议信任该连接
			t.VerifySuccess()

			//TODO: 1. 直接回调--每个客户端消息处理都在单独的协程中 2. 事件通知--发送到主逻辑协程处理
			if true == t.dis_eventqueue {
				msghandler.Handler(t.session , msg)
			}else {
				msgEvent := &def.MsgDispatchEvent{Session:t.session, Msg:msg, Handler:msghandler.Handler}
				t.ch_eventqueue <- msgEvent 
			}
		}
	}
	return true, nil
}

// --------------------------------------------------------------------------
/// @brief 使用nonblock write 有可能数据发送不完全的情况(TCP底层发送缓冲区满)
/// @brief 这里推荐使用block wirte
// --------------------------------------------------------------------------
func (t *TcpConnTask) write(data []byte) (bool, error) {
	len_total, len_write := len(data), 0
	for i:=0; i < 10; i++ {
		//t.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))	// 设置写超时, Nonblock write
		nleft := len(data)
		wlen, err := (t.conn).Write(data)	// block write
		len_write += wlen
		if err != nil {
			t.quit()
			return false, err
		}
		if wlen < nleft { 
			data = data[wlen:nleft]
			continue
		}
		break
	}

	// TODO:阻塞Write理论上不会出现数据发送不完全
	if len_write < len_total {
		errlog := fmt.Sprintf("sid [%d] [%s] write not send all data !!! want:%d real:%d", t.id, t.name, len_total, len_write)
		log.Error(errlog)
		//t.wbufleft.Reset();
		//t.wbufleft.Write(data[len_write:])
		//time.Sleep(time.Millisecond*1)		// slowdown
	}
	//log.Info("sid[%d] real send data len=%d", t.id, len_write)
	return true, nil
}


func (t *TcpConnTask) isNilSession() bool {
	if t.session == nil || reflect.ValueOf(t.session).IsNil() {
		return true
	}
	return false
}

// --------------------------------------------------------------------------
/// @brief 导出接口
///
/// @param 
// --------------------------------------------------------------------------
func (t *TcpConnTask) GetSession() def.IBaseNetSession {
	return t.session
}

// 验证通过
func (t *TcpConnTask) VerifySuccess() {
	if atomic.LoadInt32(&t.legality.VerifyState) == def.ConnVerifying {
		atomic.StoreInt32(&t.legality.VerifyState, def.ConnVerifySuccess)
		log.Info("sid[%d] conn verify ok", t.id)
	}
}

func (t *TcpConnTask) IsVerify() bool {
	return atomic.LoadInt32(&t.legality.VerifyState) == def.ConnVerifySuccess
}


// --------------------------------------------------------------------------
/// @brief OnConnect和OnClose有2中方式回调 
//	1.在MonitorLoop协程直接回调(多线程)
//	2.通过事件发送到主逻辑协程处理
// --------------------------------------------------------------------------
func (t *TcpConnTask) OnClose() {
	if t.isNilSession() { 
		return
	}

	// 连接成功才需要回调OnClose接口
	if true == t.dis_eventqueue {
		t.cb_event.OnClose(t.session)
	}else {
		msgEvent := &def.NetCloseEvent{Session:t.session, Handler:t.cb_event.OnClose}
		t.ch_eventqueue <- msgEvent
	}
}

func (t *TcpConnTask) OnConnect() {
	if t.isNilSession() {
		return
	}

	if true == t.dis_eventqueue {
		t.cb_event.OnConnect(t.session)
	}else {
		msgEvent := &def.NetConnectEvent{Session:t.session, Handler:t.cb_event.OnConnect}
		t.ch_eventqueue <- msgEvent
	}
}

