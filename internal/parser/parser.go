package parser

import (
	"fmt"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
)

type Parser interface {
	Parse(s string) (item.Item, error)
}

func New(conf config.ParserConfig) (Parser, error) {
	switch conf.ParserType {
	case "json":
		return newJsonParser(conf.Json), nil
	case "regex":
		return newRegexParser(conf.Regex)
	default:
		return jsonParser{}, fmt.Errorf("invalid parser %v", conf.ParserType)
	}
}
