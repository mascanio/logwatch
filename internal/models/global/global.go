package global

import (
	"bufio"
	"io"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mascanio/logwatch/internal/item"
	table "github.com/mascanio/logwatch/internal/models/appendable_table"
	"github.com/mascanio/logwatch/internal/parser"
)

type Model struct {
	table table.Model
	wr    *watchReader
	count int
}

func New(sc *bufio.Scanner, opts ...ModelOption) Model {
	rv := Model{
		table: table.New(
			table.WithFocused(true),
			table.WithRows(make([]table.Row, 0)),
		),
		wr: &watchReader{readChan: make(chan string)},
	}
	rv.setDefaultTableStyle()

	for _, opt := range opts {
		opt(&rv)
	}

	return rv
}

type watchReader struct {
	sc       *bufio.Scanner
	readChan chan string
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
		item, err := parser.Parse(s)
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
		case "?":
			preH := lipgloss.Height(m.table.HelpView())
			m.table.Help.ShowAll = !m.table.Help.ShowAll
			postH := lipgloss.Height(m.table.HelpView())
			// Magic that I don't understand that works
			if preH < postH {
				m.table.SetHeight(m.table.Height() - preH + postH - 4)
			} else {
				m.table.SetHeight(m.table.Height() + preH - postH + 2)
			}
			return m, nil
		}
	case watchLineReaded:
		item := item.Item(msg)
		r := table.Row{
			item.Time.Format(time.TimeOnly),
			item.Level.String(),
			item.Msg,
		}
		m.table.AppendRow(r)
		m.count++
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
		m.table.Resize(msg.Width-4,
			msg.Height-2-lipgloss.Height(m.table.HelpView())-2, 2)
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) View() string {
	return baseStyle.Render(m.table.View()) + "\n" + m.table.HelpView() + "\n"
}
