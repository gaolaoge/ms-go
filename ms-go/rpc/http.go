package rpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MsHttpsClient struct {
	client *http.Client
}

func NewHttpClient() *MsHttpsClient {
	// Transport 用于请求分发，协程安全 支持连接池
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       50 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 2 * time.Second, // 握手超时时间
		},
	}

	return &MsHttpsClient{client: client}
}

func (c *MsHttpsClient) Get(url string, args map[string]any) ([]byte, error) {
	// GET 请求参数 url?
	if args != nil && len(args) > 0 {
		url = url + "?" + c.toValues(args)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHanlde(request)
}

func (c *MsHttpsClient) PostForm(url string, args map[string]any) ([]byte, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *MsHttpsClient) PostJSON(url string, args map[string]any) ([]byte, error) {
	marshal, _ := json.Marshal(args)
	request, err := http.NewRequest("POST", url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *MsHttpsClient) Response(req *http.Request) ([]byte, error) {
	return c.responseHandle(req)
}

func (c *MsHttpsClient) responseHandle(request *http.Request) ([]byte, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		err := errors.New(fmt.Sprintf("MSHTTP_ERROR: response.StatusCode = %d", response.StatusCode))
		return nil, err
	}

	reader := bufio.NewReader(response.Body)
	defer response.Body.Close()

	buf := make([]byte, 127)
	var body []byte

	for {
		n, err := reader.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		body = append(body, buf[:n]...)
	}

	return body, nil
}

func (c *MsHttpsClient) toValues(args map[string]any) string {
	if args == nil || len(args) == 0 {
		return ""
	}
	//var str strings.Builder
	//for k, v := range args {
	//	if str.Len() > 0 {
	//		str.WriteString("&")
	//	}
	//	str.WriteString(fmt.Sprintf("%s=%s", k, v))
	//}
	//return str.String()

	params := url.Values{}
	for k, v := range args {
		params.Set(k, fmt.Sprintf("%v", v))
	}
	return params.Encode()
}
