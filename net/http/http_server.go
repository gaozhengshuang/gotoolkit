/// @file http_server.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-12-01

package http
import (
	"time"
	"io/ioutil"
	"net/http"
	"gitee.com/jntse/gotoolkit/log"
	"gitee.com/jntse/gotoolkit/net/define"
	"gitee.com/jntse/gotoolkit/util"
)


// --------------------------------------------------------------------------
/// @brief Http Response Handler 包装器
// --------------------------------------------------------------------------
type HttpResponseHandleWarpper struct {
	handler	def.HttpResponseHandle
}

func NewHttpResponseHandleWarpper(cb def.HttpResponseHandle) *HttpResponseHandleWarpper {
	return &HttpResponseHandleWarpper{handler:cb }
}

// 注意：ServeHTTP会在一个新的协程里调用，如果想要单线程处理可以将数据push到主线程chan中
func (h *HttpResponseHandleWarpper) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//log.Trace("########################################################")
	//log.Trace("r=%#v", r)
	//log.Trace("########################################################")
	//log.Trace("URL=%#v", r.URL)
	//log.Trace("URL.path=%#v", r.URL.Path)
	//log.Trace("URL.RawQuery=%#v", r.URL.RawQuery)
	//log.Trace("###########################")
	//log.Trace("RequestURI=%#v", r.RequestURI)
	//log.Trace("Method=%#v", r.Method)
	//log.Trace("Local=%#v", r.Host)
	//log.Trace("Remote=%#v", r.RemoteAddr)
	//log.Trace("Header=%#v", r.Header)
	//log.Trace("########################################################")
	
	if h.handler == nil {
		log.Error("must regist http response handle func")
		return
	}

	body , err := ioutil.ReadAll(r.Body)
	if err != nil { 
		log.Info("http request err=%s",err)
		return
	}

	defer util.RecoverPanic(nil, nil)
	w.Header().Set("HttpRemoteAddress", r.RemoteAddr)	// 为header设置RemoteAddr参数
	w.Header().Set("HttpListenSerivce", r.Host)			// 为header设置
	h.handler(w, r.URL.Path, r.URL.RawQuery, body)
	w.Header().Del("HttpRemoteAddress")					// 删除RemoteAddr参数
	w.Header().Del("HttpListenSerivce")
}


// --------------------------------------------------------------------------
/// @brief 
// --------------------------------------------------------------------------
//type http.Handler interface {
//	ServeHTTP(ResponseWriter, *Request)
//}  

type HttpListener struct {
	server 	*http.Server
	handler http.Handler
	conf	def.HttpListenConf
	start 	bool
	Quit 	bool
	ch_Quit	chan int32
}

func NewHttpListener(conf def.HttpListenConf) *HttpListener {
	return &HttpListener{conf:conf}
}

func (h *HttpListener) Name() string {
	return h.conf.Name
}

func (h* HttpListener) Host() *def.NetHost {
	return &h.conf.Host
}

func (h *HttpListener) Init(cb def.HttpResponseHandle) {

	h.ch_Quit = make(chan int32, 1)			// nonblock

	// 路由handler (不同的URL可以由不同的handler处理)
	mux := &http.ServeMux{}
	h.handler = mux
	mux.Handle("/", NewHttpResponseHandleWarpper(cb))
	//h.RegistResponseHandler("/sendgmcmd", NewHttpResponseHandleWarpper(HttpResponseHandleWarpperGmCmd))
	//h.RegistResponseHandler("/mail", NewHttpResponseHandleWarpper(HttpResponseHandleWarpperMail))
	//h.RegistResponseHandler("/onlinereport", NewHttpResponseHandleWarpper(HttpResponseHandleWarpperOnlineReport))
	//h.RegistResponseHandler("/notice", NewHttpResponseHandleWarpper(HttpResponseHandleWarpperNotice))

	// 简单handler
	//h.handler = http.FileServer(http.Dir("/home/ecs-user/gopath/src/testhttp"))	// 文件服务器
	//h.handler = NewHttpResponseHandleWarpper(cb)

	h.server  = &http.Server {
		Addr:           h.Host().String(),
		Handler:        h.handler,	// 如果Handler是nil可以使用ServerMux路由
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (h *HttpListener) RegistResponseHandler(pattern string, wrapper http.Handler) {
	mux, ok := h.handler.(*http.ServeMux)
	if ok == false { return }
	mux.Handle(pattern, wrapper)	// 注册到ServeMux中
}

func (h *HttpListener) Start() bool {
	if h.start == true { return false }
	go func() {
		h.start = true
		if h.conf.Https == true {
			re := h.server.ListenAndServeTLS(h.conf.Cert, h.conf.CertKey)		// block
			log.Error("HttpListener'[%s]' Quit ListenAndServe [%v]...", h.Name(), re)
		}else {
			re := h.server.ListenAndServe()		// block
			log.Error("HttpListener'[%s]' Quit ListenAndServe [%v]...", h.Name(), re)
		}
		h.Quit = true
		h.ch_Quit <- 1
	}()

	protocol := "http"
	if h.conf.Https { protocol = "https" }

	// 等待http监听是否成功
	timerwait := time.NewTimer(time.Millisecond * 20)
	select {
	case <-h.ch_Quit:
		log.Error("HttpListener'%s' listen'%s://%s' fail", h.Name(), protocol, h.Host().String())
		return false
	case <-timerwait.C:
		log.Info("HttpListener'%s' listen'%s://%s' ok", h.Name(), protocol, h.Host().String())
		//go h.MainLoop()
		break
	}
	return true
}

func (h *HttpListener) ShutDown() {
	if h.start == false { return }
	if h.Quit == true { return }
	h.server.Close()
	//h.server.Shutdown(context.Background())
	<-h.ch_Quit
}

//func (h *HttpListener) MainLoop() {
//	for {
//		time.Sleep(time.Millisecond*1)
//	}
//}



