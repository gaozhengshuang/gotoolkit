/// @file define.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package def
import (
	"fmt"
	"strings"
	"time"
	"net/http"
	_"sync/atomic"
)

// --------------------------------------------------------------------------
/// @brief 通用定义, 首字母大写导出，小写def内部使用
// --------------------------------------------------------------------------
const (
	KDafaultWriteQueueSize = 1000
	KServerWriteQueueSize  = 100000
	KServerRecvQueueSize = 100000
	KReadBufferSize 	= 128 * 1024
	KWriteBufferSize 	= 128 * 1024
	KFrameDispatchNum	= 1000
	KLegalityVerifyTimeOut = 10
)

const (
	KCmdRDMaxSize    	= 4096			// 单次收包大小 4k
	KCmdWRMaxSize    	= 1024			// 单次发包大小 1k
	KCmdRBufMaxSize		= 32 * 1024		// 接收缓冲区初始容量, 64k
	KCmdWBufMaxSize		= 32 * 1024		// 发送缓冲区初始容量, 64k
	KKilobyteSize		= 1024			// 1K Bytes
	KMegabyteSize		= 1048576		// 1M Bytes
	KEventQueueMaxSize 	= 100000		// 要足够大，否则队列满了后主线程消息处理效率很低
	KSlowRbufSize		= 1048576		// 1M Bytes，rbuf大于这个大小，readloop降速
	KTcpKeepAlivePeriod	= 30 * time.Second	// Tcp内部心跳间隔,秒
)


/// --------------------------------------------------------------------------
/// @brief 连接状态
// --------------------------------------------------------------------------
const (
	Unconnected = 1	// 未连接
	Connectting = 2	// 正在连接 -- not used
	Connected = 3	// 连接成功
	Closing = 4		// 正在关闭
	Closed = 5		// 完全关闭
)

/// --------------------------------------------------------------------------
/// @brief Task behavior -- 连接行为 
// --------------------------------------------------------------------------
const (
	Uninitialized = 0	// uninitialized conn
	Acceptor 	= 1		// accpet  conn
	Connector	= 2		// connect conn
)

// MsgHandler proto消息handler处理回调
type MsgHandler func(session IBaseNetSession, msg interface{})

// Http Response Hanlder
type HttpResponseHandle func(w http.ResponseWriter, urlpath string, rawquery string, body []byte)

//
type SendHandler			func(msg interface{}) bool
//type OnCloseHandler		func(userdata interface{})
//type OnEstablishedHandler	func(func(userdata interface{}))


// --------------------------------------------------------------------------
/// @brief Net Session 基础接口
// --------------------------------------------------------------------------
type IBaseNetSession interface {
	Init()
	Id() int
	Name() string
	SendCmd(msg interface{}) bool
	Close()
	SetUserDefData(udata interface{})
	UserDefData() interface{}
	DelayClose(delay int64)
	LocalIp() string
	LocalPort() int
	RemoteIp() string
	RemotePort() int
	VerifySuccess()
	IsVerify() bool
}

type HttpSession struct {
}
func (h* HttpSession) Init() {
}


// --------------------------------------------------------------------------
/// @brief Host
// --------------------------------------------------------------------------
type NetHost struct {
	Ip      string  `json:"ip"`
	Port    int     `json:"port"`
}

func NewNetHost(host string) *NetHost {
	ip , port, newhost := "", 0, strings.Replace(host, ":", " ", -1)
	fmt.Sscanf(newhost, "%s %d", &ip, &port)
	return &NetHost{ip, port}
}

func (n *NetHost) String() string {
	return fmt.Sprintf("%s:%d" ,n.Ip ,n.Port)
}

func IpPortKey(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}


// --------------------------------------------------------------------------
/// @brief 连接合法性 枚举
// --------------------------------------------------------------------------
const (
	ConnVerifying = 1		// 验证中
	ConnVerifyFailed = 2	// 验证失败
	ConnVerifySuccess = 3		// 验证完成
	ConnVerifyExclude = 4	// 不用验证
	)

type TcpConnLegality struct {
	VerifyState int32
	TimeOut int64
}

func (t* TcpConnLegality) Init() {
	t.VerifyState = ConnVerifyExclude
	t.TimeOut = 0
}

type IBaseConnTask interface {
	//Quit()
	VerifySuccess()
	GetSession() IBaseNetSession
	OnConnect()
	OnClose()
}


type INetWork interface {

	GenerateTaskId() int32
	OnSessionClose(conn IBaseConnTask)
	OnSessionEstablished(conn IBaseConnTask)

	DelTcpConnector(name string)
	DelUdpConnector(name string)
	DelWsConnector(name string)

	EventBaseCb() IBaseNetCallback
	EventQueue() EventChan
}


