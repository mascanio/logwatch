package config

import "github.com/pelletier/go-toml/v2"

type Config struct {
	Fields []Field
	Parser ParserConfig
}

type Field struct {
	Name, Type string
	Width      int
	Flex       bool
}

type ParserConfig struct {
	ParserType string
	Json       ParserJson
	Regex      ParserRegex
}

type ParserRegex struct {
	Regex  string
	Fields []struct {
		Name, Type string
	}
}

type ParserJson struct {
	Fields []struct {
		JsonKey, Type string
	}
}

func ParseConfig(doc []byte) (Config, error) {
	var rv Config
	err := toml.Unmarshal(doc, &rv)
	if err != nil {
		return Config{}, err
	}

	return rv, nil
}
