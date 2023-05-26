package ms_go

import (
	"log"
	"net"
	"strings"
	"time"
)

/*
go 标准日志库足以提供日常功能，这里只做1些优化，如添加颜色， 做日志分级。
日志中间件，用于打印一些请求信息，
*/

type LoggingConfig struct{}

func LoggingWithConfig(conf LoggingConfig, next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		if ctx.R.Method == "OPTIONS" {
			next(ctx)
			return
		}
		start := time.Now()
		next(ctx)
		stop := time.Now()
		latency := stop.Sub(start)
		path := ctx.R.URL.Path
		raw := ctx.R.URL.RawQuery
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		clientIP := net.ParseIP(ip)
		method := ctx.R.Method
		statusCode := ctx.StatusCode

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[MSGO] Logging: %v | %3d | %13v | %15s | %-7s %#v",
			stop.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

func Logging(next HandleFunc) HandleFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}
