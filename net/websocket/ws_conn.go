/// @file ws_conn.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package ws
import (
	"net"
	"fmt"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/ringbuf"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/codec"
	"time"
	"sync"
	"sync/atomic"
	_"encoding/binary"
	_"reflect"
	"net/url"
	"net/http"
	"github.com/gorilla/websocket"
	"crypto/tls"

)



// --------------------------------------------------------------------------
/// @brief websocket连接对象
/// @TODO WsConnTask golang是强类型语言，protobuf中的int是int32
/// 导致代码中很多地方需要强制类型装换 int->int32, int32->int，有时间将代码中的int使用int32代替
/// @TODO 全部梳理一遍代码，所有的chan都要在一个协程中关闭清理，多线程下极容易产生隐藏bug
// --------------------------------------------------------------------------

// 基础属性
type WsConnTaskBase struct {
	ip				string
	port			int
	id				int
	name			string
	state			int32
	conn			*websocket.Conn	// net core
	behavior 		int32			// accept/connect
	msgtype			int32			// websocket message types are defined in RFC 6455
}

// 读写Loop
type WsConnTaskRWloop struct {
	//rbuf			[]byte			// read buffer
	rbuf			*ringbuf.Buffer	// read buffer，环形缓冲区
	ch_writelen		int32			// 发送队列大小
	ch_write		chan interface{}// TODO: send buffer, send/monitor/外部都会访问(需要加锁)
	ch_quitwloop 	chan int		// 通知退出 write loop
	ch_quitrloop 	chan int		// 通知退出 read loop
	syn_wloop		sync.WaitGroup	// 等待write协程退出
	syn_rloop		sync.WaitGroup	// 等待read协程退出
}

// 额外属性
type WsConnTaskExtra struct {
	parser			codec.IBaseParser		// 消息解析器
	netcore			def.INetWork
	legality		def.TcpConnLegality
	session			def.IBaseNetSession	// *WsSession 暴露外部使用
	cb_event		def.IBaseNetCallback
	ch_eventqueue 	def.EventChan
	dis_eventqueue	bool
	userdata		interface{}		// 用户自定义数据
	svrchannel		bool			// 服务器间task
}


type WsConnTask struct {
	WsConnTaskBase
	WsConnTaskRWloop
	WsConnTaskExtra
}


// --------------------------------------------------------------------------
/// @brief 非导出方法
// --------------------------------------------------------------------------
// 新建实例
func newWsConnTask(base *WsConnTaskBase, verify bool, svrchannel bool) (*WsConnTask) {
	objconn := new(WsConnTask)
	objconn.WsConnTaskBase = *base		// TODO: shallow copy(浅拷贝)
	objconn.id = 0
	objconn.state = def.Unconnected
	objconn.msgtype = websocket.BinaryMessage

	var w_queuelen int32 = def.KDafaultWriteQueueSize
	if svrchannel { w_queuelen = def.KServerWriteQueueSize }
	objconn.ch_writelen = w_queuelen					// TODO: ch_write len 太小很快就会写满, 导致阻塞 
	objconn.ch_write = make(chan interface{}, w_queuelen)
	//objconn.rbuf = make([]byte, 0, def.KCmdRBufMaxSize)
	objconn.rbuf = ringbuf.NewBuffer(def.KCmdRBufMaxSize)
	objconn.ch_quitwloop = make(chan int, 1)
	objconn.ch_quitrloop = make(chan int, 1)
	objconn.legality.Init()
	if verify == true {
		objconn.legality.TimeOut = time.Now().Unix() + def.KLegalityVerifyTimeOut
		objconn.legality.VerifyState = def.ConnVerifying
	}
	objconn.session = newWsSession(objconn)
	objconn.svrchannel = svrchannel
	return objconn
}

func (w *WsConnTask) Init(parser codec.IBaseParser, netcore def.INetWork, dis_eventqueue bool) {
	w.parser = parser
	w.netcore = netcore
	w.id = int(netcore.GenerateTaskId())
	w.cb_event = netcore.EventBaseCb()
	w.ch_eventqueue = netcore.EventQueue()
	w.dis_eventqueue = dis_eventqueue
}

// 绑定用户数据，以函数指针方式暴露接口
func (w *WsConnTask) setUserDefdata(udata interface{})	{
	w.userdata = udata
}


// LocalAddr 获取本地地址
func (w *WsConnTask) localAddr() net.Addr {
	if w.conn != nil { return (w.conn).LocalAddr() }
	return nil
}


// RemoteAddr 获取远端地址
func (w *WsConnTask) remoteAddr() net.Addr {
	if w.conn != nil { return (w.conn).RemoteAddr() }
	return nil
}

// 关闭退出 TODO: 有可能此时read和write报错通知MonitorLoop开始走清理，导致chan monitorloop被清理掉
// 解决方案: 不使用 chan monitorloop，改为使用Closing状态来结束task
func (w *WsConnTask) quit() {
	stat := atomic.LoadInt32(&w.state)
	if stat == def.Closed || stat == def.Closing {
		//log.Error("sid[%d] is 'Closed' or 'Closing' stat[%d]", w.id, stat)
		return
	}

	atomic.StoreInt32(&w.state, def.Closing)
}

// 最后调用，释放所有协程和资源
func (w *WsConnTask) cleanup() {
	w.quitwriteLoop()	// 退出wloop
	w.waitWloopQuit()	// 等待wloop
	w.quitreadLoop()		// 退出rloop
	w.waitRloopQuit()	// 等待rloop
	w.netcore.OnSessionClose(w)

	//
	atomic.StoreInt32(&w.state, def.Closed)
	//if w.conn != nil { w.conn.Close() }
	w.conn = nil
	w.rbuf = nil
	w.session = nil

	close(w.ch_write)
	w.ch_write = nil
	close(w.ch_quitwloop)
	w.ch_quitwloop = nil
	close(w.ch_quitrloop)
	w.ch_quitrloop = nil

}


// 重置状态准备重连
// TODO: 不一定需要重新make变量，频繁断开可能会导致内存爆炸,频繁GC等问题
func (w *WsConnTask) reset() {
	w.conn = nil
	atomic.StoreInt32(&w.state, def.Unconnected)

	w.ch_write = make(chan interface{}, w.ch_writelen)
	//w.rbuf = make([]byte, 0, def.KCmdRBufMaxSize)
	if w.rbuf == nil {  w.rbuf = ringbuf.NewBuffer(def.KCmdRBufMaxSize) }
	w.rbuf.Reset()

	w.ch_quitwloop = make(chan int, 1)
	w.ch_quitrloop = make(chan int, 1)
}


// 连接
func (w *WsConnTask) connect(https bool) bool {
	host := fmt.Sprintf("%s:%d",w.ip, w.port)

	url_server := url.URL{Scheme: "ws", Host: host, Path:"/ws_handler"} 
	wsDialer := websocket.DefaultDialer
	if https == true {
		url_server = url.URL{Scheme: "wss", Host: host, Path:"/ws_handler"} 
		tslconfig := &tls.Config{InsecureSkipVerify: true}	// 忽略证书验证
		wsDialer = &websocket.Dialer{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tslconfig }
	}

	conn, _, err := wsDialer.Dial(url_server.String(), nil)
	if err != nil {
		log.Error("sid[%d] can't connect to [%s] err=%s", w.id, host, err.Error())
		return false
	}

	w.conn = conn
	atomic.StoreInt32(&w.state, def.Connected)
	w.syn_wloop.Add(1)
	w.syn_rloop.Add(1)
	w.startWritereadLoop()
	w.session = newWsSession(w)
	return true
}

// accept一个连接
func (w *WsConnTask) accpeted() {
	atomic.StoreInt32(&w.state, def.Connected)
	w.syn_wloop.Add(1)
	w.syn_rloop.Add(1)
	w.session = newWsSession(w)
	w.startWritereadLoop()
}

// TODO: 	使用通道来作为发送缓冲区，和传统缓冲区有差别
// 			0.外部协程访问
// 			1.以包作为单位，而非连续内存
//			2.致命问题: 如果ch_write通道满了，会阻塞发送操作所在协程，例如主线程
//			3.一般来说一个玩家发送队列不会积累太多消息包，发生了2情况说明这个玩家网络已经延迟很长了
//			4.致命问题: 例如发送一个1024byte包没有一次性发送完毕，不支持把剩下未发送的内容加入到发送队列最前面
//			5.最终结论: 添加一个传统的发送缓冲区
func (w *WsConnTask) sendCmd(msg interface{}) bool {
	if atomic.LoadInt32(&w.state) != def.Connected {
		log.Error("sid[%d] conn is 'Closed' and 'send' fail", w.id)
		return false
	}

	msgnum := int32(len(w.ch_write))
	if msgnum >= w.ch_writelen {
		w.quit()
		panic(fmt.Sprintf("sid:'%d' ch_write is full, force close!!!", w.id))
	}

	w.ch_write <- msg
	return true
}

func (w *WsConnTask) waitRloopQuit() {
	w.syn_rloop.Wait()
}

func (w *WsConnTask) waitWloopQuit() {
	w.syn_wloop.Wait()
}

// 启动收发Loop
func (w *WsConnTask) startWritereadLoop() {
	go w.writeLoop()
	go w.readLoop()
}


func (w* WsConnTask) quitreadLoop() {
	if w.conn != nil { w.conn.Close(); }	// 使用阻塞read需要Close()来触发read内部报错解除阻塞状态，退出readloop
	w.ch_quitrloop <- 1
}


func (w* WsConnTask) quitwriteLoop() {
	w.ch_quitwloop <- 1
}


func (w *WsConnTask) writeLoop()	{
	defer util.RecoverPanic(w.ProcessPanic, nil)
	defer w.syn_wloop.Done()
	//defer log.Info("sid[%d] writeLoop Quit", w.id)
	//log.Info("enter writeLoop[%d]", util.GetRoutineID())

	// TODO: 这里不使用原子读应该也ok，担心影响效率
	for w.state == def.Connected {
		if w.svrchannel == true {
			time.Sleep(time.Microsecond*1)
		}else {
			time.Sleep(time.Millisecond*20)
		}

		select {
		case <-w.ch_quitwloop:
			return
		case msg, open := <-w.ch_write:
			if open == false {
				log.Info("sid[%d] ch_write has been Closed", w.id)
				return
			}

			data , ok := w.parser.PackMsg(msg)
			if ok == false {
				log.Fatal("ProtoMsg PackMsg Fail")
				return
			}

			//w.wbuf = append(w.wbuf, data...)
			_, err := w.write(data)
			if err != nil {
				log.Error("sid[%d] WriteError:'%s'", w.id, err)
				return
			}
		}
	}
}

func (w* WsConnTask) readLoop() {
	defer w.syn_rloop.Done()
	//defer log.Info("sid[%d] readLoop Quit", w.id)
	//log.Info("enter readLoop[%d]", util.GetRoutineID())
	defer util.RecoverPanic(w.ProcessPanic, nil)

	// TODO: 这里不使用原子读应该也ok，担心影响效率
	for w.state == def.Connected {
		if w.svrchannel == true {
			time.Sleep(time.Microsecond*1)
		}else {
			time.Sleep(time.Millisecond*20)
		}
		select {
		case <-w.ch_quitrloop:
			return
		default:
			_, err := w.read()
			if err != nil	{
				log.Error("sid[%d] ReadError:%s", w.id, err)
				return
			}
		}
	}
}

func (w *WsConnTask) read()  (bool, error) {
	
	for i:=0; i < 10; i++ {
		//w.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1))	// 设置读超时, nonblock read
		msgtype, msg, err := w.conn.ReadMessage()		// 如果conn已经时效了，多次调用会"panic: repeated read on failed websocket connection"
		if err != nil || w.msgtype != int32(msgtype) {
			if nerr, convertok := err.(net.Error); convertok && nerr.Timeout() {  return true , nil }
			w.quit()
			return false, err
		}

		// TODO: rbuf无限增大，有风险导致内存耗尽
		//w.rbuf = append(w.rbuf, msg...)
		w.rbuf.Write(msg)
		//log.Trace("ReadFrom: sid[%d] rbuf[%d %d]", w.id, w.rbuf.Idle(), w.rbuf.Len())

		// all unpack done
		for ;; {
			msg, msghandler, err := w.parser.UnPackMsg(w.rbuf) 
			if err != nil {
				w.quit()
				break
			}

			if msg == nil || msghandler == nil {
				break
			}

			// 成功解析协议信任该连接
			w.VerifySuccess()

			//TODO: 1. 直接回调--每个客户端消息处理都在单独的协程中 2. 事件通知--发送到主逻辑协程处理
			if true == w.dis_eventqueue {
				msghandler.Handler(w.session , msg)
			}else {
				msgEvent := &def.MsgDispatchEvent{Session:w.session, Msg:msg, Handler:msghandler.Handler}
				w.ch_eventqueue <- msgEvent 
			}
		}
	}
	
	return true, nil
}

// --------------------------------------------------------------------------
/// @brief 使用nonblock write 有可能数据发送不完全的情况(TCP底层发送缓冲区满)
/// @brief 这里推荐使用block wirte
// --------------------------------------------------------------------------
func (w *WsConnTask) write(data []byte) (bool, error) {
	
	//w.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))	// 设置写超时, Nonblock write
	err := w.conn.WriteMessage(int(w.msgtype), data)		// block write
	if err != nil {
		w.quit()
		return false, err
	}

	//log.Info("sid[%d] real send data len=%d", w.id, len_write)
	return true, nil
}

// TODO：从逻辑上规避去掉反射检查nil，效率会提升很多
func (w *WsConnTask) isNilSession() bool {
	if w.session == nil /*|| reflect.ValueOf(w.session).IsNil()*/ {
		return true
	}
	return false
}


// --------------------------------------------------------------------------
/// @brief 导出接口
///
/// @param 
// --------------------------------------------------------------------------
// 验证通过
func (w *WsConnTask) VerifySuccess() {
	if atomic.LoadInt32(&w.legality.VerifyState) == def.ConnVerifying {
		atomic.StoreInt32(&w.legality.VerifyState, def.ConnVerifySuccess)
		//log.Info("sid[%d] conn verify ok", w.id)
	}
}

func (w *WsConnTask) IsVerify() bool {
	return atomic.LoadInt32(&w.legality.VerifyState) == def.ConnVerifySuccess
}

func (w *WsConnTask) GetSession() def.IBaseNetSession {
	return w.session
}

func (w *WsConnTask) ProcessPanic(data interface{}) {
	log.Fatal("taskid=%d state=%d", w.id, w.state)
	w.quit()
}

// --------------------------------------------------------------------------
/// @brief OnConnect和OnClose有2中方式回调 
//	1.在MonitorLoop协程直接回调(多线程)
//	2.通过事件发送到主逻辑协程处理
// --------------------------------------------------------------------------
func (w *WsConnTask) OnClose() {
	if w.isNilSession() {
		return
	}

	if true == w.dis_eventqueue {
		w.cb_event.OnClose(w.session)
	}else {
		msgEvent := &def.NetCloseEvent{Session:w.session, Handler:w.cb_event.OnClose}
		w.ch_eventqueue <- msgEvent
	}
}

func (w *WsConnTask) OnConnect() {
	if w.isNilSession() {
		return
	}

	if true == w.dis_eventqueue {
		w.cb_event.OnConnect(w.session)
	}else {
		msgEvent := &def.NetConnectEvent{Session:w.session, Handler:w.cb_event.OnConnect}
		w.ch_eventqueue <- msgEvent
	}
}

