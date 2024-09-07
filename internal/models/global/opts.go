package global

import (
	"bufio"

	"github.com/charmbracelet/lipgloss"
	table "github.com/mascanio/logwatch/internal/models/appendable_table"
	"github.com/mascanio/logwatch/internal/parser"
)

type ModelOption func(*Model)

func WithTableStyle(s table.Styles) ModelOption {
	return func(m *Model) {
		m.table.SetStyles(s)
	}
}

func WithScanner(sc *bufio.Scanner) ModelOption {
	return func(m *Model) {
		m.wr.sc = sc
	}
}

func WithParser(parser parser.Parser) ModelOption {
	return func(m *Model) {
		m.wr.parser = parser
	}
}

func (m *Model) setDefaultTableStyle() {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)
}
