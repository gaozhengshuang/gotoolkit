/// @file udp_listener.go
/// @brief 不完整实现(缺少包编号排序和丢包重发)
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package udp
import (
	"net"
	_"fmt"
	"time"
	"sync"
	"sync/atomic"
	"strings"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net/codec"
	"gitee.com/jntse/gotoolkit/net/define"
)


type UdpListener struct {
	conf		def.UdpListenConf
	running		bool
	locker		sync.Mutex
	//listener	*net.TCPListener	// 
	listener	*net.UDPConn
	udpConnSet  map[int]*UdpConnTask
	udpConnSetAddr map[string]*UdpConnTask
	parser		codec.IBaseParser
	blacklist	map[string]int		// 黑名单
	netcore     def.INetWork
	rbuftmp		[]byte			// TODO: 临时read buffer
}

func NewUdpListener(conf def.UdpListenConf, parser codec.IBaseParser, netcore def.INetWork) *UdpListener {
	info := new(UdpListener)
	info.conf = conf
	info.running = false
	info.listener = nil;
	info.udpConnSet = make(map[int]*UdpConnTask)
	info.udpConnSetAddr = make(map[string]*UdpConnTask)
	info.parser = parser
	info.netcore = netcore
	info.rbuftmp = make([]byte, def.KCmdRMaxSize)		// Read()需要size不是0 buffer，否则返回0
	return info
}


// --------------------------------------------------------------------------
/// @brief 导出函数
// --------------------------------------------------------------------------
func (u* UdpListener) Init() {

}

func (u* UdpListener) Start() bool {
	u.running = true
	if u.listen() == false {
		log.Error("UdpListener'%s' listen'%s' fail", u.Name(), u.Host().String())
		return false
	}
	log.Info("UdpListener'%s' listen'%s' ok", u.Name(), u.Host().String())
	go u.run()
	return true
}

// Listener 结束
func (u* UdpListener) ShutDown() {
	if u.running == false { return }
	u.running = false
	u.listener.SetDeadline(time.Now())	// accept立刻超时，退出阻塞状态
}

func (u* UdpListener) Name() string {
	return u.conf.Name
}

func (u* UdpListener) Host() *def.NetHost {
	return &u.conf.Host
}

// --------------------------------------------------------------------------
/// @brief 非导出函数
// --------------------------------------------------------------------------
func (u* UdpListener) lock() {
	u.locker.Lock()
}

func (u* UdpListener) unlock() {
	u.locker.Unlock()
}

func (u* UdpListener) listen() bool {

	// 使用UdpListener
	udpAddr, err_resolve := net.ResolveUDPAddr("udp", u.Host().String())
	if err_resolve != nil {
		log.Error("'%s' ResolveUDPAddr error:%v", u.Name(), err_resolve)
		return false
	}

	// udp conn
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Error("ListenUDP:'%s' error:%v", u.Name(), err.Error())
		return false
	}

	u.listener = conn
	//u.listener.SetWriteBuffer(def.KWriteBufferSize)
	//u.listener.SetReadBuffer(def.KReadBufferSize)
	return true
}

// accpet 协程
func (u* UdpListener) run() {
	defer u.cleanListener()
	for ;; {
		time.Sleep(time.Millisecond*1)

		if u.running == false {
			log.Info("Listener'%s'准备退出...", u.Name())
			break;
		}

		// Udp Read
		len, addr, err := u.listener.ReadFromUDP(u.rbuftmp)
		if err != nil {
			log.Error("Listener'%s' ReadFromUDP Error[%s]", u.Name(), err)
			continue
		}
		task := u.getConnTask(addr)

		// 派发到task
		task.pushRecvData(u.rbuftmp[:len])
	}
}

// 清理Listener
func (u* UdpListener) cleanListener() {

	// 关闭监听
	if (u.listener != nil) {
		u.listener.Close()
		u.listener = nil
	}

	// 清理所有task
	for _, udpconn := range u.udpConnSet {
		udpconn.Quit()
	}

	u.udpConnSetAddr = make(map[string]*UdpConnTask)
	log.Info("Listener '%s' 清理完毕", u.Name())
}


func (u* UdpListener) addConn(udpconn* UdpConnTask) {
	u.lock()
	u.udpConnSet[udpconn.id] = udpconn
	u.udpConnSetAddr[udpconn.raddr.String()] = udpconn
	u.unlock()
}


func (u* UdpListener) delConn(udpconn* UdpConnTask) {
	u.lock()
	delete(u.udpConnSet, udpconn.id)
	delete(u.udpConnSetAddr, udpconn.raddr.String())
	u.unlock()
}

func (u *UdpListener) getConnTask(addr *net.UDPAddr) *UdpConnTask {
	u.lock()
	conn, find := u.udpConnSetAddr[addr.String()]
	u.unlock()
	
	if find == true {
		return conn
	}

	return u.onAccept(addr)
}

// UdpConnTask监视协程, 非Loop
func (u* UdpListener) udpConnMonitorLoop(udpconn *UdpConnTask) {
	defer log.Info("sid[%d] '%s' Quit udpConnMonitorLoop", udpconn.id, udpconn.name)
	ticker100ms := time.NewTicker(time.Millisecond * 100)
	defer ticker100ms.Stop()
	for ;; {
		time.Sleep(time.Millisecond*100)
		stat := atomic.LoadInt32(&udpconn.state)
		if stat == def.Closed || stat == def.Closing {
			if ( stat == def.Closing ) {
				u.delConn(udpconn); 
				udpconn.cleanup() 
			}
			return
		}

		// 'select chan' 起到Sleep作用
		select {
		case <-ticker100ms.C:
			if stat == def.Connected { u.TaskLegalityCheck(udpconn) }	// 连接合法性验证
			break
		}
	}
}

// 验证 socket 合法性(only accept socket)
func (u *UdpListener) TaskLegalityCheck(udpconn* UdpConnTask) {
	if udpconn.behavior != def.Acceptor {
		return 
	}

	if udpconn.legality.VerifyState == def.ConnVerifySuccess || udpconn.legality.VerifyState == def.ConnVerifyExclude {
		return 
	}

	if udpconn.legality.TimeOut <= time.Now().Unix() {
		log.Info("sid[%d] unverified in [%d] seconds", udpconn.id, def.KLegalityVerifyTimeOut)
		atomic.StoreInt32(&udpconn.legality.VerifyState, def.ConnVerifyFailed)
		udpconn.Quit()
		return 
	}
}

// 建立新连接回调
func (u* UdpListener) onAccept(addr *net.UDPAddr) *UdpConnTask {

	// TCPConn专属设置接口
	//udpcon, ok := conn.(*net.UDPConn)
	//if ok == true {
	//	udpcon.SetKeepAlive(true)
	//	udpcon.SetKeepAlivePeriod(def.KTcpKeepAlivePeriod)	// 不要太长
	//	udpcon.SetNoDelay(true)
	//	udpcon.SetWriteBuffer(def.KWriteBufferSize)
	//	udpcon.SetReadBuffer(def.KReadBufferSize)
	//}

	// 作为玩家task发送队列大小1000，作为服务器task发送队列要大的多
	var w_queuelen int32 = def.KDafaultWriteQueueSize
	if u.conf.SvrChannel { w_queuelen = def.KServerWriteQueueSize }

	// make a new name
	name := "Task" + strings.TrimSuffix(u.conf.Name, "Listener")
	taskbase := &UdpConnTaskBase{ip:"", port:0, name:name, conn:nil, udplisten:u.listener, behavior:def.Acceptor}
	udpconn := newUdpConnTask(taskbase, w_queuelen, u.conf.Verify!=0, u.conf.SvrChannel)
	udpconn.Init(u.parser, u.netcore, u.conf.DisEventQueue)
	udpconn.setRemoteAddr(addr)
	log.Info("Listener'%s' accept new conn sid[%d] '%s' w_qlen'%d'", u.Name(), udpconn.id, addr, w_queuelen)
	udpconn.accpeted()
	go u.udpConnMonitorLoop(udpconn)
	u.addConn(udpconn)
	udpconn.netcore.OnSessionEstablished(udpconn)
	return udpconn
}

