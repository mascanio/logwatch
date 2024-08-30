package global

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	table   table.Model
	wr      *watchReader
	count   int
	freezed bool
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

type watchLineReaded string
type watchEof struct{}
type watchErr error
type watchChanClosed struct{}

func (r *watchReader) watchForLine() tea.Cmd {
	return func() tea.Msg {
		defer close(r.readChan)
		for r.sc.Scan() {
			r.readChan <- r.sc.Text()
		}
		if err := r.sc.Err(); err != nil && err != io.EOF {
			return watchErr(err)
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
		return watchLineReaded(s)
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.wr.watchForLine(),
		m.wr.waitForLine(),
	)
}

func (m Model) helpHeight() int {
	s := m.table.HelpView()
	return strings.Count(s, "\n")
}

func (m Model) isMoveMsg(msg tea.KeyMsg) bool {
	km := m.table.KeyMap
	return key.Matches(msg, km.LineUp, km.LineDown, km.PageUp, km.PageDown,
		km.HalfPageUp, km.HalfPageDown, km.GotoTop, km.GotoBottom)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.isMoveMsg(msg) {
			m.freezed = true
		} else {
			switch msg.String() {
			case "c":
				m.table.GotoBottom()
				m.freezed = false
			case "q", "ctrl+c":
				return m, tea.Quit
			case "?":
				preH := m.helpHeight()
				m.table.Help.ShowAll = !m.table.Help.ShowAll
				postH := m.helpHeight()
				// Magic that I don't understand that works
				if preH < postH {
					m.table.SetHeight(m.table.Height() - preH + postH - 4)
				} else {
					m.table.SetHeight(m.table.Height() + preH - postH + 2)
				}
				return m, nil

			}
		}
	case watchLineReaded:
		m.count++
		newRows := m.table.Rows()
		count := strconv.Itoa(m.count)
		newRows = append(newRows, table.Row{count, string(msg)})
		m.table.SetRows(newRows)
		if !m.freezed {
			m.table.MoveDown(1)
		}
		return m, m.wr.waitForLine()
	case watchChanClosed:
		// log("Chan closed")
		break
	case watchEof:
		// log("EOF")
		break
	case watchErr:
		panic(msg)
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 2 - m.helpHeight() - 2)
		newCols := m.table.Columns()
		newCols[1].Width = m.table.Width() - newCols[0].Width - 2
		m.table.SetColumns(newCols)
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
