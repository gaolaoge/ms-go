package log

import (
	"fmt"
	"time"
)

type TextFormatter struct {
}

func (t *TextFormatter) Format(param *LoggingFormatParam) string {
	now := time.Now()
	fieldsString := ""
	if param.LoggerFields != nil {
		for k, v := range param.LoggerFields {
			fieldsString += fmt.Sprintf("%s=%v ", k, v)
		}
	}
	if param.IsColor == true {
		levelColor := t.LevelColor(param.Level)
		msgColor := t.MsgColor(param.Level)
		return fmt.Sprintf(
			"%s[msgo]%s %s%v%s | level=%s%s%s | msg=%s%#v%s %s \n",
			yellow,
			reset,
			blue,
			now.Format("2006-01-02 15:04:05.000"),
			reset,
			levelColor,
			param.Level.Level(),
			reset,
			msgColor,
			param.Msg,
			reset,
			fieldsString,
		)
	}
	return fmt.Sprintf(
		"[msgo] %v | level=%s | msg=%#v %s \n",
		now.Format("2006-01-02 15:04:05.000"),
		param.Level.Level(),
		param.Msg,
		fieldsString,
	)
}

func (t TextFormatter) LevelColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}

func (t TextFormatter) MsgColor(level LoggerLevel) string {
	switch level {
	case LevelDebug:
		return blue
	case LevelInfo:
		return green
	case LevelError:
		return red
	default:
		return cyan
	}
}
