/// @file conf.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-01

package def

// --------------------------------------------------------------------------
/// @brief 网络库配置，需要导出的字段要首字母大写
// --------------------------------------------------------------------------
type TcpListenConf struct {
	Name    string  `json:"name"`				// Listener名字
	Parser  string  `json:"parser"`				// 协议解析器
	Host    NetHost	`json:"host"`				// 监听host
	Verify	int		`json:"verify"`				// 是否校验身份
	SvrChannel bool `json:"svrchannel"`			// 服务器通道(监听服务器/客户端)
	DisEventQueue bool `json:"disable_eventqueue"`	// true: 不使用主线程事件队列，task recv协程直接进行回调

}

type TcpConnectConf struct {
	Name		string  `json:"name"`			// connector名字
	Parser		string  `json:"parser"`			// 协议解析器
	Host		NetHost	`json:"host"`			// 连接host
	Interval	int 	`json:"interval"`		// 重连间隔
	SvrChannel	bool 	`json:"svrchannel"`		// 服务器通道(自己是服务器/客户端)
	Disable		int `json:"disable"`			// 禁用配置
	DisReconnect	int	`json:"disreconnect"`	// 禁用重连
	DisEventQueue bool `json:"disable_eventqueue"`	// true: 不使用主线程事件队列，task recv协程直接进行回调
}

type HttpListenConf struct {
	Name    string  `json:"name"`				// Listener名字
	Host    NetHost	`json:"host"`				// 监听host
	Https	bool	`json:'https'`				// 使用https
	Cert	string 	`json:'cert'`				// cert证书
	CertKey string	`json:'certkey'`			// cert证书私钥
}

type WsListenConf struct {
	Name    string  `json:"name"`				// Listener名字
	Parser  string  `json:"parser"`				// 协议解析器
	Host    NetHost	`json:"host"`				// 监听host
	Verify	int		`json:"verify"`				// 是否校验身份
	SvrChannel bool	`json:"svrchannel"`			// 服务器通道(监听服务器/客户端)
	DisEventQueue bool `json:"disable_eventqueue"`	// true: 不使用主线程事件队列，task recv协程直接进行回调
	Https	bool	`json:"https"`				// 使用wss
	Cert	string 	`json:'cert'`				// cert证书
	CertKey string	`json:'certkey'`			// cert证书私钥
}

type WsConnectConf struct {
	Name		string  `json:"name"`			// connector名字
	Parser		string  `json:"parser"`			// 协议解析器
	Host		NetHost	`json:"host"`			// 连接host
	Interval	int 	`json:"interval"`		// 重连间隔
	SvrChannel 	bool 	`json:"svrchannel"`		// 服务器通道(监听服务器/客户端)
	Disable		int `json:"disable"`			// 禁用配置
	DisReconnect	int	`json:"disreconnect"`	// 禁用重连
	DisEventQueue bool `json:"disable_eventqueue"`	// true: 不使用主线程事件队列，task recv协程直接进行回调
	Https	bool	`json:"https"`				// 使用wss
	Cert	string 	`json:'cert'`				// cert证书
	CertKey string	`json:'certkey'`			// cert证书私钥
}

type UdpListenConf struct {
	Name    string  `json:"name"`				// Listener名字
	Parser  string  `json:"parser"`				// 协议解析器
	Host    NetHost	`json:"host"`				// 监听host
	Verify	int		`json:"verify"`				// 是否校验身份
	SvrChannel bool `json:"svrchannel"`			// 服务器通道(监听服务器/客户端)
	DisEventQueue bool `json:"disable_eventqueue"`	// true: 不使用主线程事件队列，task recv协程直接进行回调

}

type UdpConnectConf struct {
	Name		string  `json:"name"`			// connector名字
	Parser		string  `json:"parser"`			// 协议解析器
	Host		NetHost	`json:"host"`			// 连接host
	Interval	int 	`json:"interval"`		// 重连间隔
	SvrChannel	bool 	`json:"svrchannel"`		// 服务器通道(自己是服务器/客户端)
	Disable		int `json:"disable"`			// 禁用配置
	DisReconnect	int	`json:"disreconnect"`	// 禁用重连
	DisEventQueue bool `json:"disable_eventqueue"`	// true: 不使用主线程事件队列，task recv协程直接进行回调
}


func (conf* TcpConnectConf) SetHost(ip string, port int) {
	conf.Host = NetHost{ip, port}
}

func (conf* WsConnectConf) SetHost(ip string, port int) {
	conf.Host = NetHost{ip, port}
}

// --------------------------------------------------------------------------
/// @brief mysql 配置定义
// --------------------------------------------------------------------------
type MysqlConf struct {
	Name	string			`json:"name"`			// 唯一名字
	Enable	bool			`json:"enable"`			// 启用开关
	User 	string			`json:"user"`			// 用户
	Passwd	string			`json:"passwd"`			// 密码
	Database string			`json:"database"`		// 数据库
	Host    NetHost			`json:"host"`			// 地址断开
	MaxIdleConn int32		`json:"maxidleconn"`	// 最大闲置连接数
}

// --------------------------------------------------------------------------
/// @brief Redis配置定义
// --------------------------------------------------------------------------
type NetRedisConf struct {
	Passwd		string	`json:"passwd"`
	DB			int		`json:"db"`
	Host		NetHost	`json:"host"`
	Enable		bool	`json:"enable"`
}

type TablePathConf struct {
	Excel	string	`json:"excel"`
	Json	string	`json:"json"`
	Xml		string	`json:"xml"`
}

type NetConf struct {
	Name        	string	`json:"name"`					// 配置名
	TblPath			TablePathConf `json:"tblpath"`			// 表格配置
	EventQueueSize	int `json:"event_queuesize"`			// 事件队列大小
	//IsServer		bool `json:"isserver"`					// 是否是服务器配置
	TcpListeners   	[]TcpListenConf  `json:"listens"`		// 监听配置
	TcpConnectors  	[]TcpConnectConf `json:"connects"`		// 连接配置
	HttpListeners 	[]HttpListenConf `json:"httplistens"`	// http服务
	WsListeners   	[]WsListenConf  `json:"wslistens"`		// 监听配置
	WsConnectors  	[]WsConnectConf `json:"wsconnects"`		// 连接配置
	UdpListeners   	[]UdpListenConf  `json:"udplistens"`	// 监听配置
	UdpConnectors  	[]UdpConnectConf `json:"udpconnects"`	// 连接配置
	Redis			NetRedisConf `json:"redis"`				// redis配置
	Mysql			[]MysqlConf `json:"mysql"`				// mysql配置
}

func (conf *NetConf) FindTcpConnectConf(name string) (TcpConnectConf, bool) {
	for _, conf := range conf.TcpConnectors {
		if conf.Name == name	{
			return conf, true
		}
	}
	return TcpConnectConf{}, false
}

func (conf *NetConf) FindTcpListenConf(name string) (TcpListenConf, bool) {
	for _, conf := range conf.TcpListeners {
		if conf.Name == name	{
			return conf, true
		}
	}
	return TcpListenConf{}, false
}


func (conf *NetConf) FindWsConnectConf(name string) (WsConnectConf, bool) {
	for _, conf := range conf.WsConnectors {
		if conf.Name == name	{
			return conf, true
		}
	}
	return WsConnectConf{}, false
}

func (conf *NetConf) FindWsListenConf(name string) (WsListenConf, bool) {
	for _, conf := range conf.WsListeners {
		if conf.Name == name	{
			return conf, true
		}
	}
	return WsListenConf{}, false
}

func (conf *NetConf) FindUdpConnectConf(name string) (UdpConnectConf, bool) {
	for _, conf := range conf.UdpConnectors {
		if conf.Name == name	{
			return conf, true
		}
	}
	return UdpConnectConf{}, false
}

func (conf *NetConf) FindUdpListenConf(name string) (UdpListenConf, bool) {
	for _, conf := range conf.UdpListeners {
		if conf.Name == name	{
			return conf, true
		}
	}
	return UdpListenConf{}, false
}


