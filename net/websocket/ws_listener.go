/// @file ws_listener.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package ws
import (
	_"net"
	_"fmt"
	"time"
	"sync"
	"sync/atomic"
	"strings"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/codec"
	"net/http"
	"github.com/gorilla/websocket"
	"context"
)

type WsListener struct {
	conf		def.WsListenConf
	running		bool
	locker		sync.Mutex
	server  	*http.Server		// http service
	handler 	http.Handler		// http 路由控制
	upgrader	*websocket.Upgrader	// 升级器，升级http到websocket
	wsConnSet  	map[int]*WsConnTask
	parser		codec.IBaseParser
	blacklist	map[string]int		// 黑名单
	ch_listenserv chan int			// ListenAndServe 退出通知
	netcore		def.INetWork
	//cb        	def.IBaseNetCallback
	//ch_netevent	chan def.ISessionEvent
}

func NewWsListener(conf def.WsListenConf, parser codec.IBaseParser, netcore def.INetWork) *WsListener {
	info := new(WsListener)
	info.conf = conf
	info.running = false
	info.server = nil
	info.handler = nil
	info.upgrader = nil
	info.wsConnSet = make(map[int]*WsConnTask)
	info.parser = parser
	info.netcore = netcore
	info.ch_listenserv = make(chan int, 1)
	return info
}

type WsRootHandler struct {}
func (w* WsRootHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Error("This is WebSocket RootHandler，请使用子handler:'/ws_handler'")
}


// --------------------------------------------------------------------------
/// @brief 导出函数
// --------------------------------------------------------------------------
func (w* WsListener) Init() {

	// use default options
	w.upgrader = &websocket.Upgrader{}
	w.upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO: allow all connections by default，否则白鹭客户端检查跨域会失败，不知道为何
		return true
	}

	// 路由handler (不同的URL可以由不同的handler处理)
	mux := &http.ServeMux{}
	w.handler = mux
	mux.Handle("/ws_handler", w)
	mux.Handle("/", new (WsRootHandler))

	// new httpserver
	w.server  = &http.Server {
		Addr:           w.Host().String(),
		Handler:        w.handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (w* WsListener) Start() bool {
	w.running = true
	if w.listen() == false {
		return false
	}
	return true
}

// WsListener 结束
func (w* WsListener) ShutDown() {
	w.running = false
	w.server.Shutdown(context.Background())

	<-w.ch_listenserv
	close(w.ch_listenserv)
	w.cleanListener()
}

func (w* WsListener) Name() string {
	return w.conf.Name
}

func (w* WsListener) Host() *def.NetHost {
	return &w.conf.Host
}

// Http 底层回调
func (w* WsListener) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	if w.running == false	{
		return
	}

	// 将Http升级为webscoket
	conn, err := w.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Error("升级Http协议到webscoket失败:%v", err)
		return
	}

	w.onAccept(conn)
}

// --------------------------------------------------------------------------
/// @brief 非导出函数
// --------------------------------------------------------------------------
func (w* WsListener) lock() {
	w.locker.Lock()
}

func (w* WsListener) unlock() {
	w.locker.Unlock()
}

func (w* WsListener) listen() bool {
	go w.listenCoroutine()

	protocol := "http"
	if w.conf.Https { protocol = "https" }

	// 等待ListenAndServe是否成功
	timerwait := time.NewTimer(time.Millisecond * 100)
	select {
	case <-w.ch_listenserv:
		log.Error("WsListener'%s' listen'%s://%s' fail", w.Name(), protocol, w.Host().String())
		panic("WsListener listen fail")
	case <-timerwait.C:
		log.Info("WsListener '%s' listen'%s://%s' ok", w.Name(), protocol, w.Host().String())
		break
	}

	return true
}

// --------------------------------------------------------------------------
/// @brief TODO: ListenAndServe 被动退出，Listener不会被清理
// --------------------------------------------------------------------------
func (w* WsListener) listenCoroutine() {
	var re error
	if w.conf.Https == true {
		re = w.server.ListenAndServeTLS(w.conf.Cert, w.conf.CertKey)
	}else {
		re = w.server.ListenAndServe()
	}

	w.ch_listenserv <- 1
	log.Info("WsListener'%s' Quit ListenAndServe [%v]...", w.Name(), re)

}

// WsConnTask监视协程, 非Loop
func (w* WsListener) wsConnMonitorLoop(wsconn *WsConnTask) {
	defer log.Trace("sid[%d] '%s' Quit wsConnMonitorLoop", wsconn.id, wsconn.name)
	ticker100ms := time.NewTicker(time.Millisecond * 100)
	defer ticker100ms.Stop()
	for ;; {
		time.Sleep(time.Millisecond * 100)
		// 状态检查
		stat := atomic.LoadInt32(&wsconn.state)
		if stat == def.Closed || stat == def.Closing {
			if ( stat == def.Closing ) {
				w.delConn(wsconn); 
				wsconn.cleanup() 
			}
			return
		}
		
		// 'select chan' 起到Sleep作用
		select {
		case <-ticker100ms.C:
			if stat == def.Connected { w.TaskLegalityCheck(wsconn) }	// 连接合法性验证
			break
		}
	}
}


// 验证 socket 合法性(only accept socket)
func (w *WsListener) TaskLegalityCheck(wsconn* WsConnTask) {
	if wsconn.behavior != def.Acceptor {
		return 
	}

	if wsconn.legality.VerifyState == def.ConnVerifySuccess || wsconn.legality.VerifyState == def.ConnVerifyExclude {
		return 
	}

	if wsconn.legality.TimeOut <= time.Now().Unix() {
		log.Info("sid[%d] unverified in [%d] seconds", wsconn.id, def.KLegalityVerifyTimeOut)
		atomic.StoreInt32(&wsconn.legality.VerifyState, def.ConnVerifyFailed)
		wsconn.quit()
		return 
	}
}


// 建立新连接回调
func (w* WsListener) onAccept(conn* websocket.Conn) {

	// 作为玩家task发送队列大小1000，作为服务器task发送队列要大的多
	name := "Task" + strings.TrimSuffix(w.conf.Name, "Listener")	// construct task name
	taskbase := &WsConnTaskBase{ip:"", port:0, name:name, conn:conn, behavior:def.Acceptor}
	wsconn := newWsConnTask(taskbase, w.conf.Verify!=0, w.conf.SvrChannel)
	wsconn.Init(w.parser, w.netcore, w.conf.DisEventQueue)
	wsconn.accpeted()
	log.Info("WsListener'%s' accept new conn sid[%d] '%s' ", w.Name(), wsconn.id, wsconn.remoteAddr())
	go w.wsConnMonitorLoop(wsconn)
	w.addConn(wsconn)
	wsconn.netcore.OnSessionEstablished(wsconn)

}


// 清理WsListener
func (w* WsListener) cleanListener() {

	// 关闭监听
	if (w.server != nil) {
		//w.server.Close()
		w.server = nil
	}

	// 清理所有task
	for _, wsconn := range w.wsConnSet {
		wsconn.quit()
	}

	log.Info("WsListener '%s' 清理完毕", w.Name())
}


func (w* WsListener) addConn(wsconn* WsConnTask) {
	w.lock()
	w.wsConnSet[wsconn.id] = wsconn
	w.unlock()
}


func (w* WsListener) delConn(wsconn* WsConnTask) {
	w.lock()
	delete(w.wsConnSet, wsconn.id)
	w.unlock()
}


