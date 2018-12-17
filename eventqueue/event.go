/// @file event.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-04-23

package eventque

type IEvent interface {
	Process(ch_fback chan IEvent)
	Feedback()
}

// --------------------------------------------------------------------------
/// @brief 特定事件处理
// --------------------------------------------------------------------------
type HttpPostEventHandle func(url string, body string) (string, error)
type HttpPostEventFeedback func(resp string, err error)
type HttpPostEvent struct {
	url string 
	body string
	handler HttpPostEventHandle

	resp string
	resp_err error
	feedback HttpPostEventFeedback
}

func NewHttpPostEvent(url, body string, handler HttpPostEventHandle, feed HttpPostEventFeedback) *HttpPostEvent {
	return &HttpPostEvent{url:url, body:body, handler:handler, feedback:feed}
}       

func (h *HttpPostEvent) Process(ch_fback chan IEvent) {
	h.resp, h.resp_err = h.handler(h.url, h.body)
	ch_fback <- h
}   

func (h *HttpPostEvent) Feedback() {
	h.feedback(h.resp, h.resp_err)
}

// --------------------------------------------------------------------------
/// @brief 通用事件处理(大量使用了接口，效率这一块需要测试)
// --------------------------------------------------------------------------
type CommonEventHandle func([]interface{}) []interface{}
type CommonEventFeedback func([]interface{})
type CommonEvent struct {
	iparams []interface{}			// 事件参数
	fparams []interface{}			// 事件反馈参数
	handler CommonEventHandle		// 事件处理回调
	feedback CommonEventFeedback	// 事件反馈回调
}

func NewCommonEvent(iparams []interface{}, handler CommonEventHandle, feedback CommonEventFeedback) *CommonEvent {
	return &CommonEvent{iparams:iparams, handler:handler, feedback:feedback}
}

func (co *CommonEvent) Process(ch_fback chan IEvent) {
	co.fparams = co.handler(co.iparams)
	if co.feedback != nil {
		ch_fback <- co
	}
}

func (co *CommonEvent) Feedback() {
	if co.feedback != nil { co.feedback(co.fparams) }
}


