/// @file forward_parser.go
/// @brief 纯消息转发不做任何解包和打包
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-11-23

// --------------------------------------------------------------------------
// Package codec 
/// @brief 使用google/gogo/protobuf做协议通讯
///
/// @param 
// --------------------------------------------------------------------------

package codec

import _"strings"

import "gitee.com/jntse/gotoolkit/log"
import "gitee.com/jntse/gotoolkit/net/define"
import "gitee.com/jntse/gotoolkit/util"
import "gitee.com/jntse/gotoolkit/ringbuf"

type ForwardProtoParser struct {
	BaseParser
}

const (
	ForwardParserMsgHandlerWrapperId 	= 1
	ForwardParserMsgHandlerWrapperName = "ForwardParserMsgHandlerWrapper"
)

// --------------------------------------------------------------------------
/// @brief 导出接口
///
/// @param 
// --------------------------------------------------------------------------
func NewForwardProtoParser(name string, generator MsgIndexHandler) *ForwardProtoParser {
	parser := &ForwardProtoParser{}
	parser.name = name
	parser.Init(generator)
	addParser(parser)
	return parser
}


// deprecated now will remove in future
func (p* ForwardProtoParser) RegistSendProto(msg interface{})	{ p.RegistSendMsg(msg) }
func (p* ForwardProtoParser) RegistProtoMsg(msg interface{}, handler def.MsgHandler) { p.RegistRecvMsg(msg, handler) }

// 注册发送协议
func (p *ForwardProtoParser) RegistSendMsg(msg interface{}) {
	//msg_type := reflect.TypeOf(msg)
	//name, id := msg_type.String(), p.cmdid_generator(msg)
	//info := &MsgHandlerWrapper{Id:id, Name:name, Handler:nil}
	//p.cmd_names[name] = info
	//log.Info("regist send msg[%d:%+v]", id, msg_type)
}

// 注册接收协议
func (p* ForwardProtoParser) RegistRecvMsg(msg interface{}, handler def.MsgHandler) {
	info := &MsgHandlerWrapper{Id:ForwardParserMsgHandlerWrapperId, Name:ForwardParserMsgHandlerWrapperName, Handler:handler}
	p.cmd_ids[info.Id] = info
	log.Info("Regist ForwardProtoParser MsgHandlerWrapper")
}


// 数据编码
func (p* ForwardProtoParser) PackMsg(msg interface{}) ([]byte, bool) {
	return p.PackProtocol(msg)
}


// 数据解码 -- 获得完整包返回true
func (p* ForwardProtoParser) UnPackMsg(rbuf *ringbuf.Buffer) (msg_data interface{}, msg_handler *MsgHandlerWrapper, errmsg error) {

	cmddata, cmdid := p.PreUnpack(rbuf)
	if cmddata == nil {
		return nil, nil, nil
	}

	msg_info , ok := p.cmd_ids[cmdid]
	if ok == false {
		log.Fatal("unpack msg failed, not regist msg, cmdid=%d", cmdid)
		return nil, nil, nil
	}

	return cmddata, msg_info, nil
}


// --------------------------------------------------------------------------
/// @brief 预解包
/// @brief GuajiProtoParser, GoGoParser, JsonParser, CmdParser 需要单独实现这个方法
// --------------------------------------------------------------------------
func (p* ForwardProtoParser) PreUnpack(rbuf *ringbuf.Buffer) ([]byte, int32)	{
	buflen := rbuf.Len()
	if buflen <= 0 {
		return nil, 0
	}

	// 每次解包定量大小数据，也可以解包全部rbuf数据
	buflen = util.MinInt32(buflen, int32(2048))
	cmddata := rbuf.Read(buflen)
	return cmddata, ForwardParserMsgHandlerWrapperId
}


// --------------------------------------------------------------------------
/// @brief 打包具体实现 TODO: 如果发包速度很快，势必导致申请内存频繁GC也就频繁
/// @brief ForwardProtoParser, GoGoParser, JsonParser, CmdParser 需要单独实现这个方法
// --------------------------------------------------------------------------
func (p* ForwardProtoParser) PackProtocol(msg interface{}) ([]byte, bool) {
	msgbuf, ok := msg.([]byte)
	if ok == false {
		log.Fatal("<ForwardProtoParser> packmsg fail, msg is not '[]byte' datatype")
		return nil, false
	}

	data := make([]byte, 0, len(msgbuf))
	data = append(data, msgbuf...)
	return data, true
}

