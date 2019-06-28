/// @file proto_parser.go
/// @brief 
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-21

// --------------------------------------------------------------------------
// Package codec 
/// @brief 使用google/gogo/protobuf做协议通讯
///
/// @param 
// --------------------------------------------------------------------------
package codec

import "fmt"
import "reflect"
import _"strings"
import "encoding/binary"

import "gitee.com/jntse/gotoolkit/log"
import "gitee.com/jntse/gotoolkit/net/define"
import "gitee.com/jntse/gotoolkit/ringbuf"
import "github.com/gogo/protobuf/proto"

type ProtoParser struct {
	BaseParser
}

// --------------------------------------------------------------------------
/// @brief 导出接口
///
/// @param 
// --------------------------------------------------------------------------
func NewProtoParser(name string, generator MsgIndexHandler) *ProtoParser {
	parser := &ProtoParser{}
	parser.name = name
	parser.Init(generator)
	addParser(parser)
	return parser
}


// deprecated now in the future could be remove
func (p* ProtoParser) RegistSendProto(msg interface{})	{ p.RegistSendMsg(msg) }
func (p* ProtoParser) RegistProtoMsg(msg interface{}, handler def.MsgHandler) { p.RegistRecvMsg(msg, handler) }

// 注册发送协议
func (p *ProtoParser) RegistSendMsg(msg interface{}) {
	msg_type := reflect.TypeOf(msg)
	name, id := msg_type.String(), p.cmdid_generator(msg)
	//name = strings.TrimPrefix(name, "*")	// msg is pointer type
	info := &MsgHandlerWrapper{Id:id, Name:name, Handler:nil}
	p.cmd_names[name] = info
	log.Info("regist send msg[%d:%+v]", id, msg_type)
}

// 注册接收协议
func (p* ProtoParser) RegistRecvMsg(msg interface{}, handler def.MsgHandler) {
	msg_type := reflect.TypeOf(msg)
	//protomsg := reflect.New(msg_type).Elem().Interface()
	//if protomsg != nil { log.Info("protomsg=%+v", protomsg) }

	name, id := msg_type.String(), p.cmdid_generator(msg)
	regist_type := proto.MessageType(name)
	if regist_type == nil {
		panic(fmt.Sprintf("RegistProtoMsg('%s') fail, not a 'protobuf msg' ?", name))
	}

	info := &MsgHandlerWrapper{Id:id, Name:name, Handler:handler}
	p.cmd_ids[id] = info
	log.Info("regist recv msg[%d:%+v]", id, msg_type)
}


// 数据编码
func (p* ProtoParser) PackMsg(msg interface{}) ([]byte, bool) {
	return p.PackProtocol(msg.(proto.Message))
}


// 数据解码 -- 获得完整包返回true
func (p* ProtoParser) UnPackMsg(rbuf *ringbuf.Buffer) (msg_data interface{}, msg_handler *MsgHandlerWrapper, errmsg error) {

	cmddata, cmdid:= p.PreUnpack(rbuf)
	if cmddata == nil {
		return nil, nil, nil
	}

	msg_info , ok := p.cmd_ids[cmdid]
	if ok == false {
		log.Fatal("unpack msg failed, not regist msg, cmdid=%d", cmdid)
		return nil, nil, nil
	}

	msg_type := proto.MessageType(msg_info.Name)
	protomsg := reflect.New(msg_type.Elem()).Interface()
	//protomsg := reflect.New(msg_type).Elem().Interface()
	err := proto.Unmarshal(cmddata[cmd_header_size:], protomsg.(proto.Message))
	if err != nil {
		log.Fatal("msg Unmarshal fail, name=%s" , msg_info.Name)
		//conn.Quit()	// 议错乱，非法链接? 外挂?
		return nil, nil, fmt.Errorf("msg Unmarshal fail")
	}

	if msg_info.Handler == nil {
		errinfo := fmt.Sprintf("msg:%s handler func is nil", msg_info.Name)
		log.Fatal(errinfo)
		panic(errinfo)
		//return nil, nil, fmt.Errorf("msg handler func is nil")
	}


	//// 成功解析协议信任该连接
	//conn.VerifySuccess()

	////TODO: 1. 直接回调--每个客户端消息处理都在单独的协程中
	//// 		2. 事件通知--发送到主逻辑协程处理
	////msg_info.Handler(conn.session , protomsg)

	//msgEvent := &MsgDispatchEvent{conn.GetSession(), protomsg, msg_info.Handler}
	//conn.eventChan() <- msgEvent 

	return protomsg, msg_info, nil
}


// --------------------------------------------------------------------------
/// @brief 非导出接口
///
/// @param 
// --------------------------------------------------------------------------

// proto预解析
// --------------------------------------------------------------------------
/// @brief protobuf 预解包
/// @brief ProtoParser, GoGoParser, JsonParser, CmdParser 需要单独实现这个方法
// --------------------------------------------------------------------------
func (p* ProtoParser) PreUnpack(rbuf *ringbuf.Buffer) ([]byte, int32)	{
	buflen := rbuf.Len()
	if buflen < cmd_header_size {
		return nil, 0
	}

	cmdheader := rbuf.View(cmd_header_size)
	cmdlen := binary.LittleEndian.Uint16(cmdheader[0:cmd_len_size])
	cmdid  := binary.LittleEndian.Uint16(cmdheader[cmd_len_size:cmd_header_size])
	if int32(cmdlen) > buflen {
		return nil, 0
	}

	//cmddata := (*rbuf)[:cmdlen]
	//cmddata := rbuf.Read(int32(cmdlen))
	cmddata := rbuf.ReadByReference(int32(cmdlen))	// 使用引用方式
	//log.Info("unpack msgcmd ok, buflen=%d cmdlen=%d cmdid=%d", buflen, cmdlen, cmdid)
	return cmddata, int32(cmdid)
}

// --------------------------------------------------------------------------
/// @brief 打包具体实现 TODO: 如果发包速度很快，势必导致申请内存频繁GC也就频繁
/// @brief ProtoParser, GoGoParser, JsonParser, CmdParser 需要单独实现这个方法
// --------------------------------------------------------------------------
func (p* ProtoParser) PackProtocol(msg proto.Message) ([]byte, bool) {
	cmdid := p.ExtractCmdId(msg)
	if cmdid == 0 {
		panic(fmt.Sprintf("can't find msgcmd msg=%#v", msg))
		//return nil, false
	}

	buff, err := proto.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("cmd=%d proto.Marshal fail err=%s", cmdid, err))
		//return nil, false
	}

	cmdlen := len(buff) + cmd_header_size
	header := make([]byte, cmd_header_size)
	binary.LittleEndian.PutUint16(header[0:], uint16(cmdlen))
	binary.LittleEndian.PutUint16(header[2:], uint16(cmdid))
	data := append(header, buff...)
	//log.Printf("unpackmsg ok, len=%d data=%v", len(data), data)
	return data, true
}

func (p* ProtoParser) ExtractCmdId(msg proto.Message) int32 {
	name := proto.MessageName(msg)
	return p.CmdId(name)
}


