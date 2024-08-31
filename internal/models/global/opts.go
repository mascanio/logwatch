package global

import (
	"bufio"

	"github.com/charmbracelet/lipgloss"
	table "github.com/mascanio/logwatch/internal/models/appendable_table"
)

type ModelOption func(*Model)

func WithTableColums(cols []table.Column) ModelOption {
	return func(m *Model) {
		m.table.SetColumns(cols)
	}
}

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
