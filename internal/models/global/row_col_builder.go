package global

import (
	"time"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
	table "github.com/mascanio/logwatch/internal/models/appendable_table"
)

type rowColBuilder struct {
	config config.Config
}

func newRowColBuilder(config config.Config) rowColBuilder {
	return rowColBuilder{config}
}

func (b *rowColBuilder) GetCols() []table.Column {
	rv := make([]table.Column, 0, 10)
	for _, column := range b.config.Fields {
		rv = append(rv, table.Column{Title: column.Name, Width: column.Width})
	}
	return rv
}

func (b *rowColBuilder) Resize(cols []table.Column, windowWidth int) {
	nonFlexAcuWidth := 0
	for i := range cols {
		if !b.config.Fields[i].Flex {
			nonFlexAcuWidth += b.config.Fields[i].Width
		}
	}
	for i := range cols {
		if b.config.Fields[i].Flex {
			cols[i].Width = windowWidth - nonFlexAcuWidth
		}
	}
}

func (b *rowColBuilder) Row(item item.Item) table.Row {
	r := make(table.Row, len(b.config.Fields))
	r[0] = item.Time.Format(time.TimeOnly)
	r[1] = item.Level.String()
	for i, field := range b.config.Fields[2:] {
		r[i+2] = item.VariableFields[field.Name]
	}
	return r
}
