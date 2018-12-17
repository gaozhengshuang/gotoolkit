/// @file guaji_parser.go
/// @brief 为挂机项目特别定制 protocol parser
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-11-22

// --------------------------------------------------------------------------
// Package codec 
/// @brief 使用google/gogo/protobuf做协议通讯
///
/// @param 
// --------------------------------------------------------------------------

package codec

import "fmt"
import "reflect"
//import "strings"
import "encoding/binary"
//import "encoding/gob"
//import "bytes"

import "gitee.com/jntse/gotoolkit/log"
import "gitee.com/jntse/gotoolkit/net/define"
import "gitee.com/jntse/gotoolkit/ringbuf"
import "github.com/gogo/protobuf/proto"

type GuajiProtoParser struct {
	BaseParser
}

// --------------------------------------------------------------------------
/// @brief 导出接口
///
/// @param 
// --------------------------------------------------------------------------
func NewGuajiProtoParser(name string, generator MsgIndexHandler) *GuajiProtoParser {
	parser := &GuajiProtoParser{}
	parser.name = name
	parser.Init(generator)
	addParser(parser)
	return parser
}


// deprecated now in the future could be remove
func (p* GuajiProtoParser) RegistSendProto(msg interface{})	{ p.RegistSendMsg(msg) }
func (p* GuajiProtoParser) RegistProtoMsg(msg interface{}, handler def.MsgHandle) { p.RegistRecvMsg(msg, handler) }

// 注册发送协议
func (p *GuajiProtoParser) RegistSendMsg(msg interface{}) {
	//msg_type := reflect.TypeOf(msg)
	//name, id := msg_type.String(), p.cmdid_generator(msg)
	//info := &MsgHandler{Id:id, Name:name, Handler:nil}
	//p.cmd_ids[id] = info
	//log.Info("regist send msg[%d:%+v]", id, msg_type)
}

// 注册接收协议
func (p* GuajiProtoParser) RegistRecvMsg(msg interface{}, handler def.MsgHandle) {
	msg_type := reflect.TypeOf(msg)
	name, id := msg_type.String(), p.cmdid_generator(msg)
	info := &MsgHandler{Id:id, Name:name, Handler:handler, Type:msg_type }
	p.cmd_ids[id] = info
	log.Info("regist recv msg[%d:%+v]", id, msg_type)

}


// 数据编码
func (p* GuajiProtoParser) PackMsg(msg interface{}) ([]byte, bool) {
	return p.PackProtocol(msg)
}


// 数据解码 -- 获得完整包返回true
func (p* GuajiProtoParser) UnPackMsg(rbuf *ringbuf.Buffer) (msg_data interface{}, msg_handler *MsgHandler, errmsg error) {
	cmddata, cmdid := p.PreUnpack(rbuf)
	if cmddata == nil {
		return nil, nil, nil
	}

	msg_info , ok := p.cmd_ids[cmdid]
	if ok == false {
		log.Fatal("unpack msg failed, not regist msg, cmdid=%d cmdlen=%d", cmdid, len(cmddata))
		return nil, nil, nil
	}

	msg_type := msg_info.Type
	protomsg := reflect.New(msg_type.Elem()).Interface()
	err := protomsg.(ICmdBaseProto).Unmarshal(cmddata)
	if err != nil {
		log.Fatal("msg Unmarshal fail, name=%s" , msg_info.Name)
		//conn.Quit()	// 议错乱，非法链接? 外挂?
		return nil, nil, fmt.Errorf("msg Unmarshal fail")
	}

	return protomsg, msg_info, nil
}


// --------------------------------------------------------------------------
/// @brief 非导出接口
///
/// @param 
// --------------------------------------------------------------------------

// proto预解析
// --------------------------------------------------------------------------
/// @brief 预解包
/// @brief GuajiProtoParser, GoGoParser, JsonParser, CmdParser 需要单独实现这个方法
// --------------------------------------------------------------------------
func (p* GuajiProtoParser) PreUnpack(rbuf *ringbuf.Buffer) ([]byte, int32)	{
	buflen := rbuf.Len()
	if buflen < cmd_header_size {
		return nil, 0
	}

	cmdheader := rbuf.View(cmd_header_size)
	cmdlen := binary.LittleEndian.Uint32(cmdheader[0:cmd_header_size])
	totalsize := int32(cmd_header_size + cmdlen)
	if totalsize > buflen {
		return nil, 0
	}

	cmddata := rbuf.Read(totalsize)
	cmddata = cmddata[cmd_header_size:]		// 偏移4字节头
	cmdid := binary.LittleEndian.Uint16(cmddata[0:2])
	//cmdorder := binary.LittleEndian.Uint32(cmddata[2:6])
	return cmddata, int32(cmdid)
}

// --------------------------------------------------------------------------
/// @brief 打包具体实现 TODO: 如果发包速度很快，势必导致申请内存频繁GC也就频繁
/// @brief GuajiProtoParser, GoGoParser, JsonParser, CmdParser 需要单独实现这个方法
// --------------------------------------------------------------------------
func (p* GuajiProtoParser) PackProtocol(msg interface{}) ([]byte, bool) {
	buff, err := msg.(ICmdBaseProto).Marshal()
	if err != nil {
		log.Fatal("GuajiProtoParser PackProtocol Error[%s]", err)
		return nil, false
	}

	cmdlen := len(buff) + cmd_header_size
	header := make([]byte, cmd_header_size)
	binary.LittleEndian.PutUint32(header[0:], uint32(cmdlen - cmd_header_size))
	data := append(header, buff...)
	//log.Printf("unpackmsg ok, len=%d data=%v", len(data), data)
	return data, true
}

func (p* GuajiProtoParser) ExtractCmdId(msg proto.Message) int32 {
	name := proto.MessageName(msg)
	return p.CmdId(name)
}


