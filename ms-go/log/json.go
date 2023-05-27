package log

import (
	"encoding/json"
	"fmt"
	"time"
)

type JsonFormatter struct {
	TimeDisplay bool
}

func (j JsonFormatter) Format(param *LoggingFormatParam) string {
	if param.LoggerFields == nil {
		param.LoggerFields = make(Fields)
	}

	if j.TimeDisplay {
		now := time.Now()
		param.LoggerFields["time"] = now.Format("2006-01-02 15:04:05")
	}

	param.LoggerFields["msg"] = param.Msg
	marshal, _ := json.Marshal(param.LoggerFields)

	return fmt.Sprintf(
		"%s",
		string(marshal),
	)
}
