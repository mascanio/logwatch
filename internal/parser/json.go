package parser

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
)

func newJsonParser(config config.ParserJson) Parser {
	rv := jsonParser{config, make([]string, 0)}
	for _, e := range rv.config.Fields {
		if e.Type == "string" {
			rv.variableFields = append(rv.variableFields, e.JsonKey)
		}
	}
	return rv
}

type jsonParser struct {
	config         config.ParserJson
	variableFields []string
}

type jsonItem struct {
	item.Item
	parser *jsonParser
}

func (jp jsonParser) Parse(s string) (item.Item, error) {
	elem := jsonItem{parser: &jp}
	err := json.Unmarshal([]byte(s), &elem)
	if err != nil {
		return item.Item{}, err
	}
	return item.Item{
		Time:           elem.Time,
		Level:          elem.Level,
		VariableFields: elem.VariableFields,
	}, nil
}

func (i *jsonItem) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	i.VariableFields = make(map[string]string, 2)

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	for _, field := range i.parser.config.Fields {
		val, ok := m[field.JsonKey]
		if !ok {
			continue
		}
		switch field.Type {
		case "timestamp":
			timestamp, err := i.parser.parseTimestamp(val)
			if err != nil {
				continue
			}
			i.Time = timestamp
		case "loglevel":
			logLevel, err := i.parser.parseLogLevel(val)
			if err != nil {
				continue
			}
			i.Level = logLevel
		case "string":
			s, ok := val.(string)
			if !ok {
				continue
			}
			i.VariableFields[field.JsonKey] = s
		}
	}
	return nil
}

func (jp *jsonParser) parseLogLevel(val any) (item.ItemLogLevel, error) {
	levelS, ok := val.(string)
	if !ok {
		return 0, fmt.Errorf("invalid type for timestamp")
	}
	switch levelS {
	case "DEBUG":
		return item.ITEM_LOG_LEVEL_DEBUG, nil
	case "INFO":
		return item.ITEM_LOG_LEVEL_INFO, nil
	case "WARN":
		return item.ITEM_LOG_LEVEL_WARN, nil
	case "ERROR":
		return item.ITEM_LOG_LEVEL_ERROR, nil
	case "FATAL":
		return item.ITEM_LOG_LEVEL_FATAL, nil
	}
	return 0, fmt.Errorf("invalid loglevel %v", levelS)
}

func (jp *jsonParser) parseTimestamp(val any) (time.Time, error) {
	timestampS, ok := val.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid type for timestamp")
	}
	timestamp, err := time.Parse(time.RFC3339, timestampS)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %v for format %v",
			timestampS, time.RFC3339)
	}
	return timestamp, nil
}
