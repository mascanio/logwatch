package parser

import (
	"testing"

	"github.com/mascanio/logwatch/internal/config"
)

var configFile = `
# [timestamp]
# format =

# [loglevel]
# debug = ["DEBUG"]
# info = ["INFO"]

[[fields]]
name = "Timestamp"
type = "timestamp"

[[fields]]
name = "Level"
type = "loglevel"

[[fields]]
name = "Msg"
type = "string"

[parser]
parserType = "json"

[[parser.json.fields]]
jsonKey = "Timestamp"
type = "timestamp"

[[parser.json.fields]]
jsonKey = "Level"
type = "loglevel"

[[parser.json.fields]]
jsonKey = "Msg"
type = "string"

[[parser.json.fields]]
jsonKey = "Host"
type = "string"
`

func BenchmarkParse(b *testing.B) {
	conf, _ := config.ParseConfig(configFile)
	parser := newJsonParser(conf.Parser.Json)
	str := `{"Level":"DEBUG", "Msg":"50.374006ms", "Timestamp":"2024-09-07T10:13:57.320840167+02:00"}`
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(str)
		if err != nil {
			b.Error(err)
		}
	}

}
