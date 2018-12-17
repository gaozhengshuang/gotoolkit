/// @file ws_connector.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package ws
import (
	_"fmt"
	"time"
	"sync/atomic"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/codec"
)

type WsConnector struct {
	conf		def.WsConnectConf
	running		bool						// 断开连接仍然true，只有Shutdown设置false
	interval	int							// 重连间隔
	ticker		*util.GameTicker			// 重连定时器
	conn		*WsConnTask					// 连接对象
	netcore		def.INetWork				// 网络框架
	parser		codec.IBaseParser			// 解析器
}

func NewWsConnector(conf def.WsConnectConf, parser codec.IBaseParser, netcore def.INetWork) *WsConnector {
	connector := &WsConnector{}
	connector.parser = parser;
	connector.conf = conf
	connector.interval = conf.Interval
	connector.running = true
	connector.netcore = netcore
	if connector.interval <= 0 { connector.interval = 1 }
	ip , port, name := conf.Host.Ip, conf.Host.Port, conf.Name
	taskbase := &WsConnTaskBase{ip:ip, port:port, name:name, conn:nil, behavior:def.Connector}
	connector.conn = newWsConnTask(taskbase, false, conf.SvrChannel)
	connector.conn.Init(parser, netcore, conf.DisEventQueue)
	connector.ticker = util.NewGameTicker(time.Second * time.Duration(connector.interval), connector.tickReconnect)
	connector.ticker.Start()
	connector.Init()
	return connector
}


// --------------------------------------------------------------------------
/// @brief 公用接口
// --------------------------------------------------------------------------
func (w* WsConnector) Init() {
}

func (w* WsConnector) Start() bool {
	go w.wsConnMonitorLoop()
	return true
}

// connector 结束清理task
func (w* WsConnector) ShutDown() {
	w.running = false
	stat := atomic.LoadInt32(&w.conn.state)
	if stat == def.Closed || stat == def.Closing {
		return
	}
	w.conn.quit()
}

func (w* WsConnector) ID() int {
	return w.conn.id
}

func (w* WsConnector) Name() string {
	return w.conf.Name
}

// 禁用重连
func (w *WsConnector) IsCanReconnect() bool {
	if w.running == true && w.conf.DisReconnect == 0 {
		return true
	}
	return false
}

// --------------------------------------------------------------------------
/// @brief 非导出接口
// --------------------------------------------------------------------------
func (w* WsConnector) tickReconnect(now int64) {
	if atomic.LoadInt32(&w.conn.state) == def.Unconnected { 
		w.reconnect() 
	}
}

func (w* WsConnector) reconnect() {
	if atomic.LoadInt32(&w.conn.state) == def.Connected {
		log.Error("已经连接成功了")
		return
	}

	//log.Info("sid[%d] 重连中[%s]...", w.ID(), w.conf.Host.String())
	if w.conn.connect(w.conf.Https) {
		w.onEstablished()
		return
	}

	// 可能第一次就连接失败，禁用重连就不需要在进行连接了
	if w.IsCanReconnect() == false {
		w.ShutDown()
	}
}

func (w* WsConnector) wsConnMonitorLoop()	{
	defer log.Info("sid[%d] '%s' wsConnMonitorLoop Quit Done", w.ID(), w.Name())
	defer w.ticker.Stop()
	for {
		time.Sleep(time.Millisecond * 100)

		stat := atomic.LoadInt32(&w.conn.state)
		if stat == def.Closed || stat == def.Closing {
			if ( stat == def.Closing ) {
				w.conn.cleanup() 
			}

			if w.IsCanReconnect() == false { 	// 如果禁用重连，断开后删除connector，并退出
				w.netcore.DelWsConnector(w.Name())
				return
			}

			w.conn.reset()
			log.Info("sid[%d] '%s' Start Reconnect... ", w.ID(), w.Name())
		}

		w.ticker.Run(util.CURTIMEMS())
	}
}

// 回调
func (w* WsConnector) onEstablished() {
	log.Info("sid[%d] 连接成功[%s] w_qlen[%d]", w.ID(), w.conf.Host.String(), w.conn.ch_writelen)
	w.netcore.OnSessionEstablished(w.conn)
}


