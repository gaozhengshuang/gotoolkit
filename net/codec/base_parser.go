/// @file base_parser.go
/// @brief 
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-21

package codec
import "fmt"
import "reflect"
import "gitee.com/jntse/gotoolkit/net/define"
import "gitee.com/jntse/gotoolkit/log"
import "gitee.com/jntse/gotoolkit/ringbuf"

// parser类型
const (
	proto_parser 	= 1
	json_parser  	= 2
	gogo_parser  	= 3
	struct_parser 	= 4
)

// 包头定义
const (
	cmd_id_size     = 2;	// id  大小
	cmd_len_size    = 2;	// len 大小
	cmd_header_size = 4;	// 包头大小
)

type IBaseMsgHandler interface {
	Init()
}

// 获取msg index id
type MsgIndexHandler func(msg interface{}) int32
type BatchRegistHandler func() map[int32]*MsgHandlerWrapper  


// MsgHandlerWrapper proto消息注册信息 
type MsgHandlerWrapper struct {
	Name    string
	Id      int32
	Handler def.MsgHandler
	Type	reflect.Type
}


// --------------------------------------------------------------------------
/// @brief protocol 消息接口，struct，json
// --------------------------------------------------------------------------
type ICmdBaseProto interface {
	UId() int32
	Init()
	Marshal() ([]byte, error)
	Unmarshal(buf []byte) error
}

type ICmdJsonProto interface {
	UId() int32
	Init()
	Marshal() ([]byte, error)
	Unmarshal(buf []byte) error
}


// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
type IBaseParser interface {
	Name() string
	CmdId(string) int32
	Init(handler MsgIndexHandler)
	//PreUnpack(rbuf *[]byte) (*[]byte, int32)
	//UnPackMsg(rbuf *[]byte) (msg interface{}, handler *MsgHandlerWrapper, errmsg error)
	//PackMsg(msg interface{}) ([]byte, bool)
	PreUnpack(rbuf *ringbuf.Buffer) ([]byte, int32)
	UnPackMsg(rbuf *ringbuf.Buffer) (msg interface{}, handler *MsgHandlerWrapper, errmsg error)
	PackMsg(msg interface{}) ([]byte, bool)

	RegistProtoMsg(msg interface{}, handler def.MsgHandler)	// deprecated now in the future could be remove
	RegistRecvMsg(msg interface{}, handler def.MsgHandler)
	RegistSendMsg(msg interface{})
}

// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
type BaseParser struct {
	name string
	cmd_ids map[int32]*MsgHandlerWrapper				// just for recv(unpack) msg
	cmd_names map[string]*MsgHandlerWrapper			// just for send(pack) msg
	cmdid_generator MsgIndexHandler
}

// 初始化
func (ba *BaseParser) Init(generator MsgIndexHandler) {
	ba.cmd_ids = make(map[int32]*MsgHandlerWrapper)
	ba.cmd_names = make(map[string]*MsgHandlerWrapper)
	ba.cmdid_generator = generator
}

func (ba *BaseParser) Name() string {
	return ba.name
}

func (ba *BaseParser) CmdId(s string) int32 {
	if uid := g_SendMsgHandler.CmdId(s); uid != 0 {
		return uid
	}

	info, ok := ba.cmd_names[s]
	if ok == false {
		log.Error("not regist msg=%s", s)
		return 0
	}
	return info.Id
}

func (ba *BaseParser) RegistRecvMsg(msg interface{}, handler def.MsgHandler) { }
func (ba *BaseParser) RegistSendMsg(msg interface{}) { }


// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
type JsonCmdParser struct {
	BaseParser
}

type GoGoCmdParser struct {
	BaseParser
}

type StructCmdParser struct {
	BaseParser
}

// --------------------------------------------------------------------------
/// @brief 发送协议全局注册
// --------------------------------------------------------------------------
type GlobalSendMsgHandler struct {
	cmd_names map[string]*MsgHandlerWrapper				// just for recv(unpack) msg
}

func (g *GlobalSendMsgHandler) Init(msgs map[int32]string) {
	g.cmd_names = make(map[string]*MsgHandlerWrapper)
	for id, name := range msgs {
		g.cmd_names[name] = &MsgHandlerWrapper{Id:id, Name:name, Handler:nil}
		//log.Info("regist send msg[%d:%s]", id, name)
	}
}

func (g *GlobalSendMsgHandler) CmdId(s string) int32 {
	info, ok := g.cmd_names[s]
	if ok == false {
		log.Error("not regist msg=%s", s)
		return 0
	}
	return info.Id
}

var g_SendMsgHandler GlobalSendMsgHandler

func InitGlobalSendMsgHandler(msgs map[int32]string) {
	g_SendMsgHandler.Init(msgs)
}


// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
var g_ProtoParserSet 	map[string]IBaseParser
//var g_ProtoParserSet 	map[string]ProtoParser
//var g_JsonParserSet 	map[string]*JsonCmdParser
//var g_GoGoParserSet 	map[string]*GoGoCmdParser
//var g_StructParserSet	map[string]*StructCmdParser
func init() {
	g_ProtoParserSet 	= make(map[string]IBaseParser)
}

func GetParser(name string) IBaseParser {
	parser, ok := g_ProtoParserSet[name]
	if ok == false {
		return nil
	}
	//fmt.Printf("g_ProtoParserSet=%v\n", g_ProtoParserSet)
	return parser
}

func addParser(parser IBaseParser)	{
	name := parser.Name()
	if GetParser(name) != nil {
		panic(fmt.Sprintf("parser[%s] duplicate error", name))
	}

	g_ProtoParserSet[name] = parser
}

