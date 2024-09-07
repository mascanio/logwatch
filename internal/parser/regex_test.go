package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
)

var configFileRegex = `
[[fields]]
name = "Timestamp"
type = "timestamp"
width = 8

[[fields]]
name = "Level"
type = "loglevel"
width = 6

[[fields]]
name = "Msg"
type = "string"
width = 10
flex = true

[[fields]]
name = "Host"
type = "string"
width = 10

[parser]
parserType = "regex"

[parser.regex]
regex = '\[(?P<Timestamp>.*)\] \[(?P<Level>\w+)\] (?P<Msg>.*)'

[[parser.regex.fields]]
name = "Timestamp"
type = "timestamp"

[[parser.regex.fields]]
name = "Level"
type = "loglevel"

[[parser.regex.fields]]
name = "Msg"
type = "string"
`

func TestParseRegex(t *testing.T) {
	conf, err := config.ParseConfig([]byte(configFileRegex))
	if err != nil {
		t.Fatal(err)
	}
	parser, err := newRegexParser(conf.Parser.Regex)
	assert.NoError(t, err)
	str := `[2024-09-07T18:55:07+02:00] [DEBUG] 50.205977ms`
	i, err := parser.Parse(str)
	assert.NoError(t, err)

	assert.Equal(t, i.Level, item.ITEM_LOG_LEVEL_DEBUG)
	expectedTimestamp, _ := time.Parse(time.RFC3339, "2024-09-07T18:55:07+02:00")
	assert.Equal(t, i.Time, expectedTimestamp)
	assert.Equal(t, i.VariableFields["Msg"], "50.205977ms")
}

func BenchmarkParseRegex(b *testing.B) {
	conf, err := config.ParseConfig([]byte(configFileRegex))
	if err != nil {
		b.Fatal(err)
	}
	parser, err := newRegexParser(conf.Parser.Regex)
	if err != nil {
		b.Error(err)
	}
	str := `[2024-09-07T18:55:07+02:00] [DEBUG] 50.205977ms`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(str)
	}
}
