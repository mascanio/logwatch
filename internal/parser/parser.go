package parser

import (
	"encoding/json"

	"github.com/mascanio/logwatch/internal/item"
)

type Parser struct {
}

func New() Parser {
	return Parser{}
}

func Parse(s string) (item.Item, error) {
	var rv item.Item
	err := json.Unmarshal([]byte(s), &rv)
	if err != nil {
		return item.Item{}, err
	}
	return rv, nil
}
