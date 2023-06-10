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
	"reflect"
	"strings"
	"time"
)

const (
	HTTP  = "http"
	HTTPS = "https"
)

const (
	GET       = "GET"
	POST_FORM = "Post_form"
	POST_JSON = "Post_json"
)

type MsHttpsClient struct {
	client     *http.Client
	serviceMap map[string]MsService
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

	return &MsHttpsClient{
		client:     client,
		serviceMap: make(map[string]MsService),
	}
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
	return c.responseHandle(request)
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

type HttpConfig struct {
	Protocol string
	Host     string
	Port     int
}

func (c HttpConfig) Url(url string) string {
	return fmt.Sprintf("%s://%s:%d%s", c.Protocol, c.Host, c.Port, url)
}

type MsService interface {
	Env() HttpConfig
}

func (c *MsHttpsClient) RegisterHttpService(name string, service MsService) {
	c.serviceMap[name] = service
}

func (c *MsHttpsClient) Do(service string, method string) MsService {
	s, ok := c.serviceMap[service]
	if !ok {
		panic(fmt.Sprintf("name is %s service is not registered", service))
	}
	// 给要调用的方法复赋值
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	if t.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("service is not a pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	fieldIndex := -1
	for i := 0; i < tVar.NumField(); i++ {
		name := tVar.Field(i).Name
		if name == method {
			fieldIndex = i
			break
		}
	}
	if fieldIndex == -1 {
		panic(fmt.Sprintf("method is not found"))
	}

	// TODO 这里判断 service 实例内指定属性为nil（该url未被使用）再做赋值，
	// TODO 这个赋值动作也可以在注册服务时就循环遍历掉，免去这里的动作

	tag := tVar.Field(fieldIndex).Tag
	tagInfo := tag.Get("msrpc")
	if tagInfo == "" {
		panic(fmt.Sprintf("msrpc tag is not found"))
	}

	split := strings.Split(tagInfo, ",")
	if len(split) != 2 {
		panic(fmt.Sprintf("msrpc tag is not valid"))
	}

	methodType := split[0]
	path := split[1]
	config := s.Env()

	f := func(args map[string]interface{}) ([]byte, error) {
		if methodType == GET {
			return c.Get(config.Url(path), args)
		} else if methodType == POST_JSON {
			return c.PostJSON(config.Url(path), args)
		} else if methodType == POST_FORM {
			return c.PostForm(config.Url(path), args)
		}
		return nil, errors.New("msrpc tag method is not valid")
	}

	of := reflect.ValueOf(f)
	vVar.Field(fieldIndex).Set(of)

	return s
}
