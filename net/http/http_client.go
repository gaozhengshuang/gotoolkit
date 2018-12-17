/// @file http_client.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-12-01

package http
import (
	"fmt"
	"strings"
	"io"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	_"gitee.com/jntse/gotoolkit/log"
)

// Http请求应答返回数据结构
type HttpResponse struct {
	Code 	int			// 错误码
	Status 	string		// 错误码描述
	Body 	[]byte		// Body内容
}

// --------------------------------------------------------------------------
/// @brief 非导出接口
///
/// @param kind - "GET" or "POST"
/// @param url  - "http://47.90.41.170:27015/reportonline?game_id=5189&channel_id=1&zone_id=2&number=1"
/// @param body - ""
/// @return 
// --------------------------------------------------------------------------
func sendHttpRequest(kind string, url string, ctype string, body io.Reader) (*HttpResponse, error) {
	//
	client := &http.Client{}
	req, err := http.NewRequest(kind, url, body)
	if err != nil { return nil, err }

	// only post
	if kind == "POST" {
		req.Header.Set("Content-Type", ctype)
		//req.Header.Set("Cookie", "name=anny")
	}

	// "The client must close the response body when finished with it"
	resp, err := client.Do(req)
	if err != nil {  return nil, err }
	defer resp.Body.Close()
	
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return nil, err }
	return &HttpResponse{Code:resp.StatusCode, Status: resp.Status, Body: rbody}, nil
}


// --------------------------------------------------------------------------
/// @brief 非导出接口
///
/// @param 
/// @param error
///
/// @return 
// --------------------------------------------------------------------------
func sendHttpRequestByPropertySheet(req *http.Request) (*HttpResponse, error) {
	// "The client must close the response body when finished with it"
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {  return nil, err }
	defer resp.Body.Close()
	
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return nil, err }
	return &HttpResponse{Code:resp.StatusCode, Status: resp.Status, Body: rbody}, nil
}

// --------------------------------------------------------------------------
/// @brief 导出接口
/// @brief properties可以传空: HttpPostByProperty(url, body, nil)
///
/// 使用示例
/// url := "http://127.0.0.1:27015"
/// body := "gmcmd=addbox id=1019"
/// HttpPostByProperty(url, body, map[string]string{"Content-Type":"application/json"}})
/// properties := map[string]string{ "Content-Type":ContentType, "Authorization":AuthBase64, "Content-length":Contentlength ,"Accept":ContentType}
/// network.HttpPostByProperty(url ,body, properties)
// --------------------------------------------------------------------------
func HttpSendByProperty(kind, url, body string, properties map[string]string) (*HttpResponse, error)	{
	// make request
	req, err := http.NewRequest(kind, url, strings.NewReader(body))
	if err != nil { return nil, err }

	// set property
	for k , v := range properties {
		req.Header.Set(k, v)
	}
	//fmt.Printf("req=%#v\n", req)
	resp, err := sendHttpRequestByPropertySheet(req)
	return resp, err
}


// --------------------------------------------------------------------------
/// @brief 导出接口
/// 使用示例
/// addr := "http://127.0.0.1:27015"
/// request := "/reportonline?game_id=5189&channel_id=1&zone_id=2&number=1"
/// body := "gmcmd=addbox id=1019"
/// HttpPost(addr + request, body)
// --------------------------------------------------------------------------
func HttpPost(url string, body string) (*HttpResponse, error)	{
	//ctype := "application/json"
	//ctype := "text/plain; charset=utf-8"
	ctype := "application/x-www-form-urlencoded"
	resp, err := sendHttpRequest("POST", url, ctype, strings.NewReader(body))
	//log.Info("HttpPost body:%#v", resp)
	return resp, err
}

/// @brief 导出接口
func HttpPostByJson(url string, body string) (*HttpResponse, error)	{
	ctype := "application/json"
	resp, err := sendHttpRequest("POST", url, ctype, strings.NewReader(body))
	return resp, err
}

/// @brief 导出接口
func HttpPostByXml(url string, body string) (*HttpResponse, error)	{
	ctype := "text/xml"
	resp, err := sendHttpRequest("POST", url, ctype, strings.NewReader(body))
	return resp, err
}

// --------------------------------------------------------------------------
/// @brief 导出接口
/// 使用示例
/// addr := "http://127.0.0.1:27015"
///	request := "/sendgmcmd?action=add&game_id=5189&zone_id=2"
///	HttpGet(addr + request)
// --------------------------------------------------------------------------
func HttpGet(url string) (*HttpResponse, error)	{
	resp, err := sendHttpRequest("GET", url, "", nil)
	return resp, err
}


// --------------------------------------------------------------------------
/// @brief 
///
/// @param string
/// @param 
/// @param error
///
/// @return 
// --------------------------------------------------------------------------
func HttpsPost(url, cacert, cert, certkey, body string) (*HttpResponse, error) {
	// 加载根证书
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(cacert)
	if err != nil {
		return nil, fmt.Errorf("Read CA Cert File err:%s", err)
	}
	pool.AppendCertsFromPEM(caCrt)


	cliCrt, err := tls.LoadX509KeyPair(cert, certkey)
	if err != nil {
		return nil, fmt.Errorf("Loadx509keypair err:%s", err)
	}

	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	tr := &http.Transport {
		TLSClientConfig: &tls.Config {
			RootCAs:      pool,	// 如不指定使用默认根证书
			Certificates: []tls.Certificate{cliCrt},
		},
	}


	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil { return nil, err }

	// "The client must close the response body when finished with it"
	resp, err := client.Do(req)
	if err != nil {  return nil, err }
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return nil, err }
	return &HttpResponse{Code:resp.StatusCode, Status: resp.Status, Body: rbody}, nil
}


// --------------------------------------------------------------------------
/// @brief Https Post 请求，忽略证书验证
///
/// @param url
/// @param 
/// @param error
///
/// @return 
// --------------------------------------------------------------------------
func HttpsPostSkipVerify(url, body string) (*HttpResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil { return nil, err }

	// "The client must close the response body when finished with it"
	resp, err := client.Do(req)
	if err != nil {  return nil, err }
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return nil, err }
	return &HttpResponse{Code:resp.StatusCode, Status: resp.Status, Body: rbody}, nil
}



// --------------------------------------------------------------------------
/// @brief HttpGet请求，忽略证书验证
///
/// @param 
/// @param error
///
/// @return 
// --------------------------------------------------------------------------
func HttpsGetSkipVerify(url string) (*HttpResponse, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return nil, err }

	// "The client must close the response body when finished with it"
	resp, err := client.Do(req)
	if err != nil {  return nil, err }
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil { return nil, err }
	return &HttpResponse{Code:resp.StatusCode, Status: resp.Status, Body: rbody}, nil
}

