/// @file net.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package network
import (
	"time"
	"fmt"
	"sync"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/codec"
	"gitee.com/jntse/gotoolkit/net/tcp"
	"gitee.com/jntse/gotoolkit/net/udp"
	"gitee.com/jntse/gotoolkit/net/websocket"
	"gitee.com/jntse/gotoolkit/net/http"
)


type NetWork struct {
	netconf			def.NetConf
	tcplisteners	map[string]*tcp.TcpListener
	tcpconnectors	map[string]*tcp.TcpConnector
	tcplocker 		sync.Mutex
	httplisteners 	map[string]*http.HttpListener
	wslisteners		map[string]*ws.WsListener
	wsconnectors	map[string]*ws.WsConnector
	udplisteners	map[string]*udp.UdpListener
	udpconnectors	map[string]*udp.UdpConnector
	wslocker 		sync.Mutex
	udplocker 		sync.Mutex
	ch_main			chan string
	running			bool
	initflag		bool
	cb				def.IBaseNetCallback
	ch_eventqueue	def.EventChan
	cb_http			def.HttpResponseHandle
	def.NetSessionPool
}

func init() {
	fmt.Println("jntse/gotoolkit/net.init()")
}

// --------------------------------------------------------------------------
/// @brief 导出函数
///
/// @param 
// --------------------------------------------------------------------------

//// NetWork 是Singleton单实例对象
//var g_pNetWorkIns *NetWork = nil
//func NetWorkIns() *NetWork {
//	if g_pNetWorkIns == nil { g_pNetWorkIns = newNetWork() }
//	return g_pNetWorkIns
//}

func NewNetWork() *NetWork {
	server := new(NetWork)
	server.netconf = def.NetConf{}
	server.running = false
	server.initflag= false
	server.ch_main = make(chan string, 1)	// 非阻塞
	server.tcplisteners = make(map[string]*tcp.TcpListener)
	server.tcpconnectors = make(map[string]*tcp.TcpConnector)
	server.wslisteners = make(map[string]*ws.WsListener)
	server.wsconnectors= make(map[string]*ws.WsConnector)
	server.httplisteners = make(map[string]*http.HttpListener)
	server.udplisteners = make(map[string]*udp.UdpListener)
	server.udpconnectors = make(map[string]*udp.UdpConnector)
	server.cb_http = nil
	return server
}

func (n* NetWork) Init(netconf *def.NetConf, cb def.IBaseNetCallback) {
	n.netconf = *netconf
	n.cb = cb
	queue_size := def.KEventQueueMaxSize
	if netconf.EventQueueSize != 0 { queue_size = netconf.EventQueueSize }	// 优先使用配置
	n.ch_eventqueue = make(def.EventChan, queue_size)  // TODO: 要足够大，否则队列满后阻塞导致效率大幅降低
	n.initflag = true
	n.NetSessionPool.Init()
}

func (n* NetWork) Start() bool {
	if !n.initflag {
		log.Error("NetWork should init first")
		return false
	}

	n.running = true
	if n.startTcpListener() == false {
		return false
	}

	if n.startHttpListener() == false {
		return false
	}

	if n.startWsListener() == false {
		return false
	}

	if n.startUdpListener() == false {
		return false
	}


	if n.startTcpConnector() == false {
		return  false
	}

	if n.startWsConnector() == false {
		return  false
	}

	if n.startUdpConnector() == false {
		return  false
	}

	return true
}

func (n* NetWork) Shutdown() {
	n.running = false
	for _, listener := range n.tcplisteners {
		listener.ShutDown()
	}

	for _, connector := range n.tcpconnectors {
		connector.ShutDown()
	}

	for _, listener := range n.httplisteners {
		listener.ShutDown()
	}

	for _, listener := range n.wslisteners {
		listener.ShutDown()
	}

	for _, connector := range n.wsconnectors {
		connector.ShutDown()
	}

	for _, listener := range n.udplisteners {
		listener.ShutDown()
	}

	for _, connector := range n.udpconnectors {
		connector.ShutDown()
	}


	time.Sleep(time.Millisecond*100)
}

func (n* NetWork) IsClosed() bool {
	return !n.running
}


// --------------------------------------------------------------------------
/// @brief 非导出函数
///
/// @param 
// --------------------------------------------------------------------------
func (n* NetWork) startHttpListener() bool {
	for _ , l_conf := range n.netconf.HttpListeners {
		listener := http.NewHttpListener(l_conf)
		listener.Init(n.cb_http)		// 通用http response handler func
		if listener.Start() == false {
			return false
		}
		n.httplisteners[listener.Name()] = listener
	}

	return true
}

func (n* NetWork) startTcpListener() bool {
	for _ , l_conf := range n.netconf.TcpListeners {
		parser := codec.GetParser(l_conf.Parser)
		if parser == nil {
			log.Error("NetWork start listener fail not found Parser'%s'", l_conf.Parser)
			return false
		}

		listener := tcp.NewTcpListener(l_conf, parser, n)
		listener.Init()
		if listener.Start() == false {
			return false
		}
		n.tcplisteners[listener.Name()] = listener
		//log.Info("NetWork startTcpListener '%s' ok", l_conf.Name)
	}

	return true
}

func (n* NetWork) startTcpConnector() bool {
	for _, c_conf := range n.netconf.TcpConnectors {
		parser := codec.GetParser(c_conf.Parser)
		if parser == nil {
			log.Error("NetWork start connector fail not found Parser'%s'", c_conf.Parser)
			return false
		}

		// 配置是否被禁用
		if c_conf.Disable == 1 {
			continue	
		}

		// IsServer: 自己是服务器
		connector := tcp.NewTcpConnector(c_conf, parser, n)
		connector.Init()
		if connector.Start() == false {
			return false
		}
		n.tcpconnectors[connector.Name()] = connector
		log.Info("NetWork startTcpConnector '%s' ok", c_conf.Name)
	}
	return true
}


func (n* NetWork) startWsListener() bool {
	for _ , l_conf := range n.netconf.WsListeners {
		parser := codec.GetParser(l_conf.Parser)
		if parser == nil {
			log.Error("NetWork start wslistener fail not found Parser'%s'", l_conf.Parser)
			return false
		}

		listener := ws.NewWsListener(l_conf, parser, n)
		listener.Init()
		if listener.Start() == false {
			return false
		}
		n.wslisteners[listener.Name()] = listener
		//log.Info("NetWork startWsListener '%s' ok", l_conf.Name)
	}

	return true
}

func (n* NetWork) startWsConnector() bool {
	for _, c_conf := range n.netconf.WsConnectors {
		parser := codec.GetParser(c_conf.Parser)
		if parser == nil {
			log.Error("NetWork start wsconnector fail not found Parser'%s'", c_conf.Parser)
			return false
		}

		// 配置是否被禁用
		if c_conf.Disable == 1 {
			continue	
		}

		// IsServer: 自己是服务器
		connector := ws.NewWsConnector(c_conf, parser, n)
		connector.Init()
		if connector.Start() == false {
			return false
		}
		n.wsconnectors[connector.Name()] = connector
		log.Info("NetWork startWsConnector '%s' ok", c_conf.Name)
	}
	return true
}


func (n* NetWork) startUdpListener() bool {
	for _ , l_conf := range n.netconf.UdpListeners {
		parser := codec.GetParser(l_conf.Parser)
		if parser == nil {
			log.Error("NetWork start udp listener fail not found Parser'%s'", l_conf.Parser)
			return false
		}

		listener := udp.NewUdpListener(l_conf, parser, n)
		listener.Init()
		if listener.Start() == false {
			return false
		}
		n.udplisteners[listener.Name()] = listener
		//log.Info("NetWork startTcpListener '%s' ok", l_conf.Name)
	}

	return true
}

func (n* NetWork) startUdpConnector() bool {
	for _, c_conf := range n.netconf.UdpConnectors {
		parser := codec.GetParser(c_conf.Parser)
		if parser == nil {
			log.Error("NetWork start udp connector fail not found Parser'%s'", c_conf.Parser)
			return false
		}

		// 配置是否被禁用
		if c_conf.Disable == 1 {
			continue	
		}

		// IsServer: 自己是服务器
		connector := udp.NewUdpConnector(c_conf, parser, n)
		connector.Init()
		if connector.Start() == false {
			return false
		}
		n.udpconnectors[connector.Name()] = connector
		log.Info("NetWork startUdpConnector '%s' ok", c_conf.Name)
	}
	return true
}



// 断开并删除一个TcpConnector
func (n *NetWork) DelTcpConnector(name string) {
	n.tcplocker.Lock()
	//connector, ok := n.tcpconnectors[name]
	_, ok := n.tcpconnectors[name]
	if ok == true {
		delete(n.tcpconnectors, name)
		n.tcplocker.Unlock()
		//connector.ShutDown()
		log.Info("NetWork DelTcpConnector '%s' ok", name)
		return
	}
	n.tcplocker.Unlock()

	if ok == false {
		log.Info("NetWork DelTcpConnector not found'%s' ", name)
		return
	}
}


// 增加一个新TcpConnector，线程安全
//func (n *NetWork) AddTcpConnector(conf def.TcpConnectConf, cb def.IBaseNetCallback) bool {
func (n *NetWork) AddTcpConnector(conf def.TcpConnectConf) bool {
	parser := codec.GetParser(conf.Parser)
	if parser == nil {
		log.Error("NetWork AddTcpConnector fail not find Parser'%s'", conf.Parser)
		return false
	}

	n.tcplocker.Lock()
	if _, ok := n.tcpconnectors[conf.Name]; ok == true {
		n.tcplocker.Unlock()
		log.Error("NetWork AddTcpConnector fail connector exist conf[%s]", conf.Name)
		return false
	}
	n.tcplocker.Unlock()

	// IsServer: 自己是服务器
	connector := tcp.NewTcpConnector(conf, parser, n)
	connector.Init()
	if connector.Start() == false {
		return false
	}
	n.tcplocker.Lock()
	n.tcpconnectors[connector.Name()] = connector
	n.tcplocker.Unlock()
	log.Info("NetWork AddTcpConnector '%s' ok", conf.Name)
	return true
}



// 断开并删除一个WsConnector
func (n *NetWork) DelWsConnector(name string) {
	n.wslocker.Lock()
	//connector, ok := n.wsconnectors[name]
	_, ok := n.wsconnectors[name]
	if ok == true {
		delete(n.wsconnectors, name)
		n.wslocker.Unlock()
		//connector.ShutDown()
		log.Info("NetWork DelWsConnector '%s' ok", name)
		return
	}
	n.wslocker.Unlock()

	if ok == false {
		log.Info("NetWork DelWsConnector not found'%s' ", name)
		return
	}
}


// 增加一个新WsConnector
func (n *NetWork) AddWsConnector(conf def.WsConnectConf) bool {
	parser := codec.GetParser(conf.Parser)
	if parser == nil {
		log.Error("NetWork AddWsConnector fail not find Parser'%s'", conf.Parser)
		return false
	}

	n.wslocker.Lock()
	if _, ok := n.wsconnectors[conf.Name]; ok == true {
		n.wslocker.Unlock()
		log.Error("NetWork AddWsConnector fail connector exist conf[%s]", conf.Name)
		return false
	}
	n.wslocker.Unlock()

	// IsServer: 自己是服务器
	connector := ws.NewWsConnector(conf, parser, n)
	connector.Init()
	if connector.Start() == false {
		return false
	}

	n.wslocker.Lock()
	n.wsconnectors[connector.Name()] = connector
	n.wslocker.Unlock()
	log.Info("NetWork AddWsConnector '%s' ok", conf.Name)

	return true
}

func (n *NetWork) DelUdpConnector(name string) {
	n.tcplocker.Lock()
	_, ok := n.udpconnectors[name]
	if ok == true {
		delete(n.udpconnectors, name)
		n.tcplocker.Unlock()
		log.Info("NetWork DelUdpConnector '%s' ok", name)
		return
	}
	n.tcplocker.Unlock()

	if ok == false {
		log.Info("NetWork DelUdpConnector not found'%s' ", name)
		return
	}
}

func (n *NetWork) AddUdpConnector(conf def.UdpConnectConf) bool {
	parser := codec.GetParser(conf.Parser)
	if parser == nil {
		log.Error("NetWork AddUdpConnector fail not find Parser'%s'", conf.Parser)
		return false
	}

	n.udplocker.Lock()
	if _, ok := n.udpconnectors[conf.Name]; ok == true {
		n.udplocker.Unlock()
		log.Error("NetWork AddUdpConnector fail connector exist conf[%s]", conf.Name)
		return false
	}
	n.udplocker.Unlock()

	// IsServer: 自己是服务器
	connector := udp.NewUdpConnector(conf, parser, n)
	connector.Init()
	if connector.Start() == false {
		return false
	}
	n.udplocker.Lock()
	n.udpconnectors[connector.Name()] = connector
	n.udplocker.Unlock()
	log.Info("NetWork AddUdpConnector '%s' ok", conf.Name)
	return true
}


func (n *NetWork) SetHttpResponseHandler(cb_http def.HttpResponseHandle) {
	n.cb_http = cb_http
}

// --------------------------------------------------------------------------
/// @brief 退出条件：处理指定num事件或者队列为空
///
/// @param num 单次最大处理消息数量
/// @param timeout 超时ms，避免Dispatch消耗时间过大
/// @return processed 返回实际处理消息数量
// --------------------------------------------------------------------------
func (n *NetWork) Dispatch(num int, timeout int64) (processed int) {
	timestamp := util.CURTIMEMS()
	for i:= 0; i < num; i++ {
		select {
		case event, open := <-n.ch_eventqueue:
			if open == false {
				panic("ch_eventqueue is Closed")
			}
			event.Process()
			processed++
		default:
			return processed
		}
		if (util.CURTIMEMS() - timestamp >= timeout) && timeout > 0 {
			return processed
		}
	}
	return processed
}

func (n *NetWork) EventQueueSize() int {
	return len(n.ch_eventqueue)
}

func (n *NetWork) EventQueueCapacity() int {
	return cap(n.ch_eventqueue)
}

func (n *NetWork) SessionSize() int32 {
	return n.NetSessionPool.Size()
}

func (n *NetWork) EventBaseCb() def.IBaseNetCallback {
	return n.cb
}

func (n *NetWork) EventQueue() def.EventChan {
	return n.ch_eventqueue
}
