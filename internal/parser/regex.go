package parser

import (
	"fmt"
	"time"

	re "github.com/mascanio/regexp-named"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
)

type regexParser struct {
	config         config.ParserRegex
	regex          re.RegexpNamed
	variableFields []string
}

func newRegexParser(config config.ParserRegex) (Parser, error) {
	rv := regexParser{config: config, variableFields: make([]string, 0)}
	re, err := re.Compile(config.Regex)
	if err != nil {
		return regexParser{}, err
	}
	rv.regex = re
	for _, e := range rv.config.Fields {
		if e.Type == "string" {
			rv.variableFields = append(rv.variableFields, e.Name)
		}
	}
	return rv, nil
}

func (rp regexParser) Parse(s string) (item.Item, error) {
	m0, m := rp.regex.FindStringNamed(s)
	if m0 == "" {
		return item.Item{}, fmt.Errorf("failed to parse %v", s)
	}
	rv := item.Item{}
	rv.VariableFields = make(map[string]string, 2)

	var err error
	for _, field := range rp.config.Fields {
		switch field.Type {
		case "timestamp":
			rv.Time, err = rp.parseTimestamp(m[field.Name])
			if err != nil {
				return item.Item{}, err
			}
		case "loglevel":
			rv.Level, err = rp.parseLogLevel(m[field.Name])
			if err != nil {
				return item.Item{}, err
			}
		case "string":
			rv.VariableFields[field.Name] = m[field.Name]
		}
	}
	return rv, nil
}

func (rp *regexParser) parseLogLevel(val string) (item.ItemLogLevel, error) {
	switch val {
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
	return 0, fmt.Errorf("invalid loglevel %v", val)
}

func (rp *regexParser) parseTimestamp(val string) (time.Time, error) {
	timestamp, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp %v for format %v",
			val, time.RFC3339)
	}
	return timestamp, nil
}
