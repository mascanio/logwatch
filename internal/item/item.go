package item

import (
	"errors"
	"time"
)

type ItemLogLevel int

const (
	ITEM_LOG_LEVEL_DEBUG ItemLogLevel = iota
	ITEM_LOG_LEVEL_INFO
	ITEM_LOG_LEVEL_WARN
	ITEM_LOG_LEVEL_ERROR
	ITEM_LOG_LEVEL_FATAL
)

func (i ItemLogLevel) String() string {
	switch i {
	case ITEM_LOG_LEVEL_DEBUG:
		return "DEBUG"
	case ITEM_LOG_LEVEL_INFO:
		return "INFO"
	case ITEM_LOG_LEVEL_WARN:
		return "WARN"
	case ITEM_LOG_LEVEL_ERROR:
		return "ERROR"
	case ITEM_LOG_LEVEL_FATAL:
		return "FATAL"
	default:
		panic(errors.New("unkown log level"))
	}
}

type Item struct {
	Time           time.Time
	Level          ItemLogLevel
	VariableFields map[string]string
	Msg            string
}
