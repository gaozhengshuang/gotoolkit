/// @file udp_connector.go
/// @brief 不完整实现(缺少包编号排序和丢包重发)
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package udp
import (
	_"fmt"
	"time"
	"sync/atomic"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/util"
	"gitee.com/jntse/gotoolkit/net/codec"
	"gitee.com/jntse/gotoolkit/net/define"
)

type UdpConnector struct {
	conf		def.UdpConnectConf
	running		bool						// 断开连接仍然true，只有Shutdown设置false
	interval	int							// 重连间隔
	ticker		*util.GameTicker			// 重连定时器
	conn		*UdpConnTask				// 连接对象
	netcore		def.INetWork					// 网络框架
	parser		codec.IBaseParser					// 解析器
}

func NewUdpConnector(conf def.UdpConnectConf, parser codec.IBaseParser, netcore def.INetWork) *UdpConnector {
	connector := &UdpConnector{}
	connector.parser = parser;
	connector.conf = conf
	connector.interval = conf.Interval
	connector.running = true
	connector.netcore = netcore
	if connector.interval <= 0 { connector.interval = 1 }
	ip , port, name := conf.Host.Ip, conf.Host.Port, conf.Name
	var w_queuelen int32 = def.KDafaultWriteQueueSize
	//if isServer == true { w_queuelen = def.KServerWriteQueueSize }
	if conf.SvrChannel { w_queuelen = def.KServerWriteQueueSize }
	taskbase := &UdpConnTaskBase{ip:ip, port:port, name:name, conn:nil, udplisten:nil, behavior:def.Connector}
	connector.conn = newUdpConnTask(taskbase, w_queuelen, false, conf.SvrChannel)
	connector.conn.Init(parser, netcore, conf.DisEventQueue)
	connector.ticker = util.NewGameTicker(time.Second * time.Duration(connector.interval), connector.tickReconnect)
	connector.ticker.Start()
	connector.Init()
	return connector
}


// --------------------------------------------------------------------------
/// @brief 公用接口
// --------------------------------------------------------------------------
func (u* UdpConnector) Init() {
}

func (u* UdpConnector) Start() bool {
	go u.udpConnMonitorLoop()
	return true
}

func (u* UdpConnector) ID() int {
	return u.conn.id
}

func (u* UdpConnector) Name() string {
	return u.conf.Name
}

// 禁用重连
func (u *UdpConnector) IsCanReconnect() bool {
	if u.running == true && u.conf.DisReconnect == 0 {
		return true
	}
	return false
}

// connector 结束清理task
func (u* UdpConnector) ShutDown() {
	u.running = false
	stat := atomic.LoadInt32(&u.conn.state)
	if stat == def.Closed || stat == def.Closing {
		return
	}
	u.conn.Quit()
}

// --------------------------------------------------------------------------
/// @brief 非导出接口
// --------------------------------------------------------------------------
func (u* UdpConnector) tickReconnect(now int64) {
	if atomic.LoadInt32(&u.conn.state) == def.Unconnected { 
		u.reconnect() 
	}
}


func (u* UdpConnector) reconnect() {
	if atomic.LoadInt32(&u.conn.state) == def.Connected {
		log.Error("已经连接成功了")
		return
	}

	//log.Info("sid[%d] 重连中[%s]...", u.ID(), u.conf.Host.String())
	if u.conn.connect() {
		u.onEstablished()
		return
	}

	// 可能第一次就连接失败，禁用重连就不需要在进行连接了
	if u.IsCanReconnect() == false {
		u.ShutDown()
	}
}

func (u* UdpConnector) udpConnMonitorLoop()	{
	defer log.Info("sid[%d] '%s' udpConnMonitorLoop Quit Done", u.ID(), u.Name())
	defer u.ticker.Stop()
	for {
		time.Sleep(time.Millisecond * 100)

		stat := atomic.LoadInt32(&u.conn.state)
		if stat == def.Closed || stat == def.Closing {
			if ( stat == def.Closing ) {
				u.conn.cleanup() 
			}

			if u.IsCanReconnect() == false { 	// 如果禁用重连，断开后删除connector，并退出
				u.netcore.DelUdpConnector(u.Name())
				return
			}

			u.conn.reset()
			log.Info("sid[%d] '%s' Start Reconnect... ", u.ID(), u.Name())
		}

		u.ticker.Run(util.CURTIMEMS())
	}
}

// 回调
func (u* UdpConnector) onEstablished() {
	log.Info("sid[%d] 连接成功[%s] w_qlen[%d]", u.ID(), u.conf.Host.String(), u.conn.ch_writelen)
	u.netcore.OnSessionEstablished(u.conn)
}


