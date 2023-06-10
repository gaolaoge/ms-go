package rpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func (c *MsHttpsClient) Get(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHanlde(request)
}

func (c *MsHttpsClient) responseHanlde(request *http.Request) ([]byte, error) {
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
