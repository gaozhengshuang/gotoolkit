/// @file export.go
/// @brief 导出网络库接口
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-01-15

package network
import (
	"gitee.com/jntse/gotoolkit/net/codec"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/net/http"
	"gitee.com/jntse/gotoolkit/net/network"
)

// --------------------------------------------------------------------------
/// @brief package network 导出接口
/// @brief 类型使用type alias 导出
/// @brief 函数使用var 导出
// --------------------------------------------------------------------------
type NetWork = network.NetWork
var  NewNetWork = network.NewNetWork


// --------------------------------------------------------------------------
/// @brief package codec 导出接口
/// @brief 类型使用type alias导出
/// @brief 函数使用var 导出
// --------------------------------------------------------------------------
type ProtoParser = codec.ProtoParser		// type alias
type IBaseParser = codec.IBaseParser 		// type alias
type IBaseMsgHandler = codec.IBaseMsgHandler
type GuajiProtoParser = codec.GuajiProtoParser
type ForwardProtoParser = codec.ForwardProtoParser
type ICmdBaseProto = codec.ICmdBaseProto
type ICmdJsonProto = codec.ICmdJsonProto

var  NewProtoParser = codec.NewProtoParser
var  NewGuajiProtoParser = codec.NewGuajiProtoParser
var  NewForwardProtoParser = codec.NewForwardProtoParser
var  InitGlobalSendMsgHandler = codec.InitGlobalSendMsgHandler


// --------------------------------------------------------------------------
/// @brief package def 导出接口
/// @brief 类型使用type alias 导出
/// @brief 函数使用var 导出
// --------------------------------------------------------------------------
type IBaseNetSession = def.IBaseNetSession
type TablePathConf = def.TablePathConf
type NetConf = def.NetConf
type WsListenConf = def.WsListenConf
type TcpConnectConf = def.TcpConnectConf
type MysqlConf = def.MysqlConf
type NetHost = def.NetHost
var  NewNetHost = def.NewNetHost
var  KFrameDispatchNum = def.KFrameDispatchNum


// --------------------------------------------------------------------------
/// @brief package network 导出接口
/// @brief 类型使用type alias 导出
/// @brief 函数使用var 导出
// --------------------------------------------------------------------------
type HttpResponse = http.HttpResponse	// type alias
var  HttpsPost = http.HttpsPost
var  HttpPost = http.HttpPost
var  HttpGet = http.HttpGet
var  HttpSendByProperty = http.HttpSendByProperty
var  HttpsPostSkipVerify = http.HttpsPostSkipVerify
var  HttpsGetSkipVerify = http.HttpsGetSkipVerify


