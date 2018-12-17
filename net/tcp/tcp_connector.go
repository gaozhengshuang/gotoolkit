/// @file tcp_connector.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package tcp
import (
	_"fmt"
	"time"
	"sync/atomic"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/codec"
)

type TcpConnector struct {
	conf		def.TcpConnectConf
	running		bool						// 断开连接仍然true，只有Shutdown设置false
	interval	int							// 重连间隔
	ticker		*util.GameTicker			// 重连定时器
	conn		*TcpConnTask				// 连接对象
	netcore		def.INetWork				// 网络框架
	parser		codec.IBaseParser			// 解析器
}

func NewTcpConnector(conf def.TcpConnectConf, parser codec.IBaseParser, netcore def.INetWork) *TcpConnector {
	connector := &TcpConnector{}
	connector.parser = parser;
	connector.conf = conf
	connector.interval = conf.Interval
	connector.running = true
	connector.netcore = netcore
	if connector.interval <= 0 { connector.interval = 1 }
	ip , port, name := conf.Host.Ip, conf.Host.Port, conf.Name
	var w_queuelen int32 = def.KDafaultWriteQueueSize
	if conf.SvrChannel { w_queuelen = def.KServerWriteQueueSize }

	taskbase := &TcpConnTaskBase{ip:ip, port:port, name:name, conn:nil, behavior:def.Connector}
	connector.conn = newTcpConnTask(taskbase, w_queuelen, false, conf.SvrChannel)
	connector.conn.Init(parser, netcore, conf.DisEventQueue)
	connector.ticker = util.NewGameTicker(time.Second * time.Duration(connector.interval), connector.tickReconnect)
	connector.ticker.Start()
	connector.Init()
	return connector
}


// --------------------------------------------------------------------------
/// @brief 公用接口
// --------------------------------------------------------------------------
func (t* TcpConnector) Init() {
}

func (t* TcpConnector) Start() bool {
	go t.tcpConnMonitorLoop()
	return true
}

func (t* TcpConnector) ID() int {
	return t.conn.id
}

func (t* TcpConnector) Name() string {
	return t.conf.Name
}

// 禁用重连
func (t *TcpConnector) IsCanReconnect() bool {
	if t.running == true && t.conf.DisReconnect == 0 {
		return true
	}
	return false
}

// connector 结束清理task
func (t* TcpConnector) ShutDown() {
	t.running = false
	stat := atomic.LoadInt32(&t.conn.state)
	if stat == def.Closed || stat == def.Closing {
		return
	}
	t.conn.quit()
}

// --------------------------------------------------------------------------
/// @brief 非导出接口
// --------------------------------------------------------------------------
func (t* TcpConnector) tickReconnect(now int64) {
	if atomic.LoadInt32(&t.conn.state) == def.Unconnected { 
		t.reconnect() 
	}
}

func (t* TcpConnector) reconnect() {
	if atomic.LoadInt32(&t.conn.state) == def.Connected {
		log.Error("已经连接成功了")
		return
	}

	//log.Info("sid[%d] 重连中[%s]...", t.ID(), t.conf.Host.String())
	if t.conn.connect() {
		t.onEstablished()
		return
	}

	// 可能第一次就连接失败，禁用重连就不需要在进行连接了
	if t.IsCanReconnect() == false {
		t.ShutDown()
	}
}

func (t* TcpConnector) tcpConnMonitorLoop()	{
	defer log.Info("sid[%d] '%s' tcpConnMonitorLoop Quit Done", t.ID(), t.Name())
	defer t.ticker.Stop()
	for {
		time.Sleep(time.Millisecond * 100)

		stat := atomic.LoadInt32(&t.conn.state)
		if stat == def.Closed || stat == def.Closing {
			if ( stat == def.Closing ) {
				t.conn.cleanup() 
			}

			if t.IsCanReconnect() == false { 	// 如果禁用重连，断开后删除connector，并退出
				t.netcore.DelTcpConnector(t.Name())
				return
			}

			t.conn.reset()
			log.Info("sid[%d] '%s' Start Reconnect... ", t.ID(), t.Name())
		}

		t.ticker.Run(util.CURTIMEMS())
	}
}

// 回调
func (t* TcpConnector) onEstablished() {
	log.Info("sid[%d] 连接成功[%s] w_qlen[%d]", t.ID(), t.conf.Host.String(), t.conn.ch_writelen)
	t.netcore.OnSessionEstablished(t.conn)
}


