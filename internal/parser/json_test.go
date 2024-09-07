package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
)

var configFileJson = `
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

func TestParseJson(t *testing.T) {
	conf, err := config.ParseConfig([]byte(configFileJson))
	if err != nil {
		t.Fatal(err)
	}
	parser := newJsonParser(conf.Parser.Json)
	str := `{"Level":"DEBUG", "Msg":"50.374006ms", "Timestamp":"2024-09-07T10:13:57.320840167+02:00"}`
	i, err := parser.Parse(str)
	assert.NoError(t, err)

	assert.Equal(t, i.Level, item.ITEM_LOG_LEVEL_DEBUG)
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2024-09-07T10:13:57.320840167+02:00")
	assert.Equal(t, i.Time, expectedTimestamp)
	assert.Equal(t, i.VariableFields["Msg"], "50.374006ms")
}

func BenchmarkParseJson(b *testing.B) {
	conf, err := config.ParseConfig([]byte(configFileJson))
	if err != nil {
		b.Fatal(err)
	}
	parser := newJsonParser(conf.Parser.Json)
	str := `{"Level":"DEBUG", "Msg":"50.374006ms", "Timestamp":"2024-09-07T10:13:57.320840167+02:00"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(str)
	}
}
