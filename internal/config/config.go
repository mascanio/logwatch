package config

import (
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Fields []Field
	Parser ParserConfig
}

type Field struct {
	Name, Type string
}

type ParserConfig struct {
	ParserType string
	Json       ParserJson
}

type ParserJson struct {
	Fields []struct {
		JsonKey, Type string
	}
	variableFields []string
}

func ParseConfig(p string) (Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var rv Config
	doc, err := io.ReadAll(f)
	if err != nil {
		return Config{}, err
	}
	err = toml.Unmarshal(doc, &rv)
	if err != nil {
		return Config{}, err
	}

	return rv, nil
}
