package global

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mascanio/logwatch/internal/config"
	"github.com/mascanio/logwatch/internal/item"
	table "github.com/mascanio/logwatch/internal/models/appendable_table"
	"github.com/mascanio/logwatch/internal/models/statusbar"
	"github.com/mascanio/logwatch/internal/parser"
)

type Model struct {
	table         table.Model
	statusbar     statusbar.Model
	wr            *watchReader
	rowColBuilder rowColBuilder
	last          time.Time
	h             int
}

func New(config config.Config, opts ...ModelOption) Model {
	sb := statusbar.New()

	rowColBuilder := newRowColBuilder(config)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(false)
	rv := Model{
		table: table.New(
			table.WithColBuilder(&rowColBuilder),
			table.WithFocused(true),
			table.WithRows(make([]table.Row, 0)),
			table.WithStyles(s),
		),
		rowColBuilder: rowColBuilder,
		statusbar:     sb,
		wr: &watchReader{
			readChan: make(chan string),
		},
	}

	for _, opt := range opts {
		opt(&rv)
	}

	return rv
}

type watchReader struct {
	sc       *bufio.Scanner
	readChan chan string
	parser   parser.Parser
}

type watchLineReaded item.Item
type watchEof struct{}
type watchErr struct{ error }
type watchChanClosed struct{}
type parserError struct{ error }

func (r *watchReader) watchForLine() tea.Cmd {
	return func() tea.Msg {
		defer close(r.readChan)
		for r.sc.Scan() {
			r.readChan <- r.sc.Text()
		}
		if err := r.sc.Err(); err != nil && err != io.EOF {
			return watchErr{err}
		}
		return watchEof{}
	}
}

func (r *watchReader) waitForLine() tea.Cmd {
	return func() tea.Msg {
		s, ok := <-r.readChan
		if !ok {
			return watchChanClosed{}
		}
		item, err := r.parser.Parse(s)
		if err != nil {
			return parserError{err}
		}
		return watchLineReaded(item)
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.wr.watchForLine(),
		m.wr.waitForLine(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
			// case "?":
			// 	preH := lipgloss.Height(m.table.HelpView())
			// 	m.table.Help.ShowAll = !m.table.Help.ShowAll
			// 	postH := lipgloss.Height(m.table.HelpView())
			// 	// Magic that I don't understand that works
			// 	if preH < postH {
			// 		m.table.SetHeight(m.table.Height() - preH + postH - 4 - 1)
			// 	} else {
			// 		m.table.SetHeight(m.table.Height() + preH - postH + 2 - 1)
			// 	}
			// 	return m, nil
		}
	case watchLineReaded:
		freq := time.Since(m.last)
		m.last = time.Now()
		item := item.Item(msg)
		m.table.AppendRow(m.rowColBuilder.Row(item))
		m.statusbar.SetContent("pipe", "",
			fmt.Sprintf("%v/%v", m.table.Cursor(), len(m.table.Rows())), freq.String())
		return m, m.wr.waitForLine()
	case parserError:
		panic(msg)
	case watchChanClosed:
		log.Println("Chan closed")
	case watchEof:
		log.Println("EOF")
	case watchErr:
		panic(msg)
	case tea.WindowSizeMsg:
		m.h = msg.Height
		m.table.Resize(msg.Width, msg.Height-lipgloss.Height(m.statusbar.View()), 2)
		m.statusbar.SetSize(msg.Width)
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		baseStyle.Render(m.table.View()),
		m.statusbar.View(),
	)
}
