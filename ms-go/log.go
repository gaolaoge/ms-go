package ms_go

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
go 标准日志库足以提供日常功能，这里只做1些优化，如添加颜色， 做日志分级。
日志中间件，用于打印一些请求信息，
*/

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

var DefaultWriter io.Writer = os.Stdout

type LoggingConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LoggerFormatter = func(params *LogFormatterParams) string

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	IsDisplayColor bool // 当日志输出到文件中时无法显示颜色，「颜色标记」会以字符的形式注入到内容行中，不利于阅读
}

func (log LogFormatterParams) StatusCodeColor() string {
	code := log.StatusCode
	switch code {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

func (log LogFormatterParams) ResetColor() string {
	return reset
}

var defaultFormatter = func(params *LogFormatterParams) string {
	statusCodeColor := params.StatusCodeColor()
	reset := params.ResetColor()

	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}

	if params.IsDisplayColor {
		return fmt.Sprintf(
			"%s [msgo] %s%s %v %s|%s %3d %s| %13v | %15s | %-7s %#v \n",
			yellow,
			reset,
			blue,
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			reset,
			statusCodeColor,
			params.StatusCode,
			reset,
			params.Latency,
			params.ClientIP,
			params.Method,
			params.Path,
		)
	}
	return fmt.Sprintf(
		"[msgo] %v |%s %3d %s| %13v | %15s | %-7s %#v \n",
		params.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusCodeColor,
		params.StatusCode,
		reset,
		params.Latency,
		params.ClientIP,
		params.Method,
		params.Path,
	)
}

func LoggingWithConfig(conf LoggingConfig, next HandleFunc) HandleFunc {
	foramtter := conf.Formatter
	out := conf.out

	if foramtter == nil {
		foramtter = defaultFormatter
	}

	if out == nil {
		out = DefaultWriter
	}

	return func(ctx *Context) {
		if ctx.R.Method == "OPTIONS" {
			next(ctx)
			return
		}

		start := time.Now()
		next(ctx)
		stop := time.Now()
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(ctx.R.RemoteAddr))
		raw := ctx.R.URL.RawQuery
		path := ctx.R.URL.Path

		if raw != "" {
			path = path + "?" + raw
		}

		params := &LogFormatterParams{
			Request:        ctx.R,
			TimeStamp:      time.Now(),
			StatusCode:     ctx.StatusCode,
			Latency:        stop.Sub(start),
			ClientIP:       net.ParseIP(ip),
			Method:         ctx.R.Method,
			Path:           path,
			IsDisplayColor: true,
		}
		fmt.Fprintf(out, foramtter(params))
	}
}

func Logging(next HandleFunc) HandleFunc {
	return LoggingWithConfig(LoggingConfig{}, next)
}
