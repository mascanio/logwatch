package item

import (
	"encoding/json"
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
		panic(errors.New("Uunkown log level"))
	}
}

type Item struct {
	Time  time.Time
	Level ItemLogLevel
	Msg   string
}

func (i *Item) UnmarshalJSON(data []byte) error {
	var realJson struct {
		Level     string
		Timestamp time.Time
		Msg       string
	}

	if err := json.Unmarshal(data, &realJson); err != nil {
		return err
	}

	switch realJson.Level {
	case "DEBUG":
		i.Level = ITEM_LOG_LEVEL_DEBUG
	case "INFO":
		i.Level = ITEM_LOG_LEVEL_INFO
	case "WARN":
		i.Level = ITEM_LOG_LEVEL_WARN
	case "ERROR":
		i.Level = ITEM_LOG_LEVEL_ERROR
	case "FATAL":
		i.Level = ITEM_LOG_LEVEL_FATAL
	}
	i.Time = realJson.Timestamp
	i.Msg = realJson.Msg
	return nil
}
