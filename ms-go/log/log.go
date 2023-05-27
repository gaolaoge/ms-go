package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

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

type Fields map[string]any

type LoggerLevel int

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

func (l LoggerLevel) Level() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type LoggingFormatter interface {
	Format(param *LoggingFormatParam) string
}

type LoggingFormatParam struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
	Msg          any
}

type LoggerFormatter struct {
	Level        LoggerLevel
	IsColor      bool
	LoggerFields Fields
}

func (f *LoggerFormatter) format(msg any) string {
	now := time.Now()
	if f.IsColor == true {
		return fmt.Sprintf(
			"%s[msgo]%s %s%v%s |%s level=%s %s|%s msg=%#v %s| fields=%#v \n",
			yellow,
			reset,
			blue,
			now.Format("2006-01-02 15:04:05.000"),
			reset,
			magenta,
			f.Level.Level(),
			reset,
			green,
			msg,
			reset,
			f.LoggerFields,
		)
	}
	return fmt.Sprintf(
		"[msgo] %v | level=%s | msg=%#v | fields=%#v \n",
		now.Format("2006-01-02 15:04:05.000"),
		f.Level.Level(),
		msg,
		f.LoggerFields,
	)
}

type Logger struct {
	Formatter    LoggingFormatter
	Level        LoggerLevel
	Outs         []io.Writer
	LoggerFields Fields
}

func (l Logger) Info(content string) {
	l.Print(LevelInfo, content)
}

func (l Logger) Debug(content string) {
	l.Print(LevelDebug, content)
}

func (l Logger) Error(content string) {
	l.Print(LevelError, content)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	// 为了避免单次赋值会影响其它位置的日志调用，这里会返回一个新的 Logger 实例
	return &Logger{
		Formatter:    l.Formatter,
		Level:        l.Level,
		Outs:         l.Outs,
		LoggerFields: fields,
	}
}

func (l Logger) Print(level LoggerLevel, content string) {
	// 若实例级别大于调用级别，忽略
	if level < l.Level {
		return
	}

	params := &LoggingFormatParam{
		Level:        level,
		LoggerFields: l.LoggerFields,
		Msg:          content,
	}
	var str string
	for _, out := range l.Outs {
		if out == os.Stdout {
			params.IsColor = true
		} else {
			params.IsColor = false
		}
		str = l.Formatter.Format(params)
		fmt.Fprintf(out, str)
	}
}

func New() *Logger {
	return &Logger{}
}

// Default 默认配置
func Default() *Logger {
	logger := &Logger{
		Formatter: &TextFormatter{},
		Level:     LevelDebug,
		Outs:      []io.Writer{os.Stdout},
	}
	return logger
}

func FileWriter(filename string) io.Writer {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return file
}
