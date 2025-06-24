package utils

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"speedtest-gd/global"
	"time"
)

var defaultClient = &http.Client{
	Timeout: 10 * time.Second,
}

// HTTPGet 发送 GET 请求
// 参数：
//
//	url: 请求的目标地址
//	headers: 自定义请求头 map[string]string，如 {"User-Agent": "Test"}
//
// 返回值：
//
//	响应体的字节切片, 错误信息
func HTTPGet(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// HTTPPost 发送 POST 请求
// 参数：
//
//	url: 请求的目标地址
//	headers: 自定义请求头 map[string]string
//	data: POST 请求体 ([]byte)，例如 JSON 字符串的字节形式
//
// 返回值：
//
//	响应体的字节切片, 错误信息
func HTTPPost(url string, headers map[string]string, data []byte) ([]byte, error) {
	reqBody := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type PingResult struct {
	Agent *global.ApacheAgent
	Delay time.Duration
	Err   error
}

var (
	dnsStart, dnsDone   time.Time
	connStart, connDone time.Time
	tlsStart, tlsDone   time.Time
	firstByte           time.Time
)

func PingNode(agent *global.ApacheAgent) (time.Duration, error) {
	dialer := &net.Dialer{
		Timeout: 3 * time.Second,
	}
	start := time.Now()
	conn, err := dialer.Dial("tcp", agent.HostIP)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	return time.Since(start), nil
}
