package service

import "github.com/gaolaoge/ms-go/rpc"

type GoodsService struct {
	Find func(args map[string]interface{}) ([]byte, error) `msrpc:"GET,/goods/find"`
}

func (s GoodsService) Env() rpc.HttpConfig {
	return rpc.HttpConfig{
		Host:     "localhost",
		Port:     9002,
		Protocol: "http",
	}
}
