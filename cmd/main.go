package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mascanio/logwatch/internal/input"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var logFile *os.File

func log(s string) {
	if _, err := logFile.WriteString(s); err != nil {
		panic(err)
	}
}

type model struct {
	table table.Model
	wr    *watchReader
	count int
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

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.wr.watchForLine(),
		m.wr.waitForLine(),
	)
}

func (m model) helpHeight() int {
	s := m.table.HelpView()
	return strings.Count(s, "\n")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
	case watchLineReaded:
		m.count++
		newRows := m.table.Rows()
		count := strconv.Itoa(m.count)
		newRows = append(newRows, table.Row{count, string(msg)})
		m.table.SetRows(newRows)
		// m.table.MoveDown(1)
		return m, m.wr.waitForLine()
	case watchChanClosed:
		log("Chan closed")
	case watchEof:
		log("EOF")
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

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n" + m.table.HelpView() + "\n"
}

func main() {
	var err error
	logFile, err = tea.LogToFile("log", "debug")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "msg", Width: 70},
	}

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

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithRows(make([]table.Row, 0)),
		table.WithStyles(s),
	)

	inputStream, err := input.NewBasicPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(inputStream)
	sc := bufio.NewScanner(reader)
	sc.Split(bufio.ScanLines)

	wr := &watchReader{readChan: make(chan string), sc: sc}
	p := tea.NewProgram(
		model{table: t, wr: wr},
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		tea.WithInputTTY(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
