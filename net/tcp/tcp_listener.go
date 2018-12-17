/// @file tcp_listener.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package tcp
import (
	"net"
	_"fmt"
	"time"
	"sync"
	"sync/atomic"
	"strings"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/codec"
)


type TcpListener struct {
	conf		def.TcpListenConf
	running		bool
	locker		sync.Mutex
	//listener	net.Listener		// interface
	listener	*net.TCPListener	// 
	tcpConnSet  map[int]*TcpConnTask
	parser		codec.IBaseParser
	blacklist	map[string]int		// 黑名单
	netcore     def.INetWork
}

func NewTcpListener(conf def.TcpListenConf, parser codec.IBaseParser, netcore def.INetWork) *TcpListener {
	info := new(TcpListener)
	info.conf = conf
	info.running = false
	info.listener = nil;
	info.tcpConnSet = make(map[int]*TcpConnTask)
	info.parser = parser
	info.netcore = netcore
	return info
}


// --------------------------------------------------------------------------
/// @brief 导出函数
// --------------------------------------------------------------------------
func (t* TcpListener) Init() {

}

func (t* TcpListener) Start() bool {
	t.running = true
	if t.listen() == false {
		log.Error("TcpListener'%s' listen'%s' fail", t.Name(), t.Host().String())
		return false
	}
	log.Info("TcpListener'%s' listen'%s' ok", t.Name(), t.Host().String())
	go t.run()
	return true
}

// Listener 结束
func (t* TcpListener) ShutDown() {
	if t.running == false { return }
	t.running = false
	t.listener.SetDeadline(time.Now())	// accept立刻超时，退出阻塞状态
	//t.listener.Close()					// 可以让accept报错超时，退出阻塞状态
}

func (t* TcpListener) Name() string {
	return t.conf.Name
}

func (t* TcpListener) Host() *def.NetHost {
	return &t.conf.Host
}

// --------------------------------------------------------------------------
/// @brief 非导出函数
// --------------------------------------------------------------------------
func (t* TcpListener) lock() {
	t.locker.Lock()
}

func (t* TcpListener) unlock() {
	t.locker.Unlock()
}

func (t* TcpListener) listen() bool {

	//*/ 使用TcpListener，可以设置accpet超时
	tcpAddr, err_resolve := net.ResolveTCPAddr("tcp", t.Host().String())
	if err_resolve != nil {
		log.Error("'%s' ResolveTCPAddr error:%v", t.Name(), err_resolve)
		return false
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	/*/
	listener, err := net.Listen("tcp", t.Host().String())
	//*/
	if err != nil {
		log.Error("'%s' error listening:%s", t.Name(), err.Error())
		return false //终止程序
	}
	t.listener = listener
	return true
}

// accpet 协程
func (t* TcpListener) run() {
	defer t.cleanListener()
	for ;; {
		time.Sleep(time.Millisecond*1)

		if t.running == false {
			log.Info("Listener'%s'准备退出...", t.Name())
			break;
		}

		tcplistener := t.listener
		//tcplistener.SetDeadline(time.Now().Add(time.Millisecond * 1))	// 设置超时，'accept' will nonblock
		conn, err := tcplistener.AcceptTCP()	// use block mode
		if err != nil {
			if nerr, convertok := err.(net.Error); convertok && nerr.Timeout() { continue }
			log.Error("Listener'%s' error accept:%v", t.Name(), err.Error())
			continue
		}

		t.onAccept(conn)
	}
}


// TcpConnTask监视协程, 非Loop
func (t* TcpListener) tcpConnMonitorLoop(tcpconn *TcpConnTask) {
	defer log.Info("sid[%d] '%s' Quit tcpConnMonitorLoop", tcpconn.id, tcpconn.name)
	ticker100ms := time.NewTicker(time.Millisecond * 100)
	defer ticker100ms.Stop()
	for ;; {
		time.Sleep(time.Millisecond * 100)
		stat := atomic.LoadInt32(&tcpconn.state)
		if stat == def.Closed || stat == def.Closing {
			if ( stat == def.Closing ) {
				t.delConn(tcpconn); 
				tcpconn.cleanup() 
			}
			return
		}

		// 'select chan' 起到Sleep作用
		select {
		case <-ticker100ms.C:
			if stat == def.Connected { t.TaskLegalityCheck(tcpconn) }	// 连接合法性验证
			break
		}
	}
}

// 验证 socket 合法性(only accept socket)
func (t *TcpListener) TaskLegalityCheck(tcpconn* TcpConnTask) {
	if tcpconn.behavior != def.Acceptor {
		return 
	}

	if tcpconn.legality.VerifyState == def.ConnVerifySuccess || tcpconn.legality.VerifyState == def.ConnVerifyExclude {
		return 
	}

	if tcpconn.legality.TimeOut <= time.Now().Unix() {
		log.Info("sid[%d] unverified in [%d] seconds", tcpconn.id, def.KLegalityVerifyTimeOut)
		atomic.StoreInt32(&tcpconn.legality.VerifyState, def.ConnVerifyFailed)
		tcpconn.quit()
		return 
	}
}

// 建立新连接回调
func (t* TcpListener) onAccept(conn net.Conn) {

	// TCPConn专属设置接口
	tcpcon, ok := conn.(*net.TCPConn)
	if ok == true {
		tcpcon.SetKeepAlive(true)
		tcpcon.SetKeepAlivePeriod(def.KTcpKeepAlivePeriod)	// 不要太长
		tcpcon.SetNoDelay(true)
		tcpcon.SetWriteBuffer(def.KWriteBufferSize)
		tcpcon.SetReadBuffer(def.KReadBufferSize)
	}

	// 作为玩家task发送队列大小1000，作为服务器task发送队列要大的多
	var w_queuelen int32 = def.KDafaultWriteQueueSize
	if t.conf.SvrChannel { w_queuelen = def.KServerWriteQueueSize }

	// make a new name
	name := "Task" + strings.TrimSuffix(t.conf.Name, "Listener")
	taskbase := &TcpConnTaskBase{ip:"", port:0, name:name, conn:tcpcon, behavior:def.Acceptor}
	tcpconn := newTcpConnTask(taskbase, w_queuelen, t.conf.Verify!=0, t.conf.SvrChannel)
	//tcpconn := newTcpConnTask("", 0, name, def.Acceptor , w_queuelen, tcpcon, t.conf.Verify != 0, t.conf.SvrChannel)
	tcpconn.Init(t.parser, t.netcore, t.conf.DisEventQueue)
	log.Info("Listener'%s' accept new conn sid[%d] '%s' w_qlen'%d'", t.Name(), tcpconn.id, tcpconn.remoteAddr(), w_queuelen)
	tcpconn.accpeted()
	go t.tcpConnMonitorLoop(tcpconn)
	t.addConn(tcpconn)
	tcpconn.netcore.OnSessionEstablished(tcpconn)
}


// 清理Listener
func (t* TcpListener) cleanListener() {

	// 关闭监听
	if (t.listener != nil) {
		t.listener.Close()
		t.listener = nil
	}


	// 清理所有task
	for _, tcpconn := range t.tcpConnSet {
		tcpconn.quit()
	}

	log.Info("Listener '%s' 清理完毕", t.Name())
}


func (t* TcpListener) addConn(tcpconn* TcpConnTask) {
	t.lock()
	t.tcpConnSet[tcpconn.id] = tcpconn
	t.unlock()
}


func (t* TcpListener) delConn(tcpconn* TcpConnTask) {
	t.lock()
	delete(t.tcpConnSet, tcpconn.id)
	t.unlock()
}


