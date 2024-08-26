//go:build !windows
// +build !windows

package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	readChan chan any
}

type watchLineReaded string
type watchEof struct{}
type watchErr error
type watchChanClosed struct{}

func watchForLine(sc *bufio.Scanner) chan any {
	readChan := make(chan any)
	go func() {
		defer close(readChan)
		for sc.Scan() {
			log("reading\n")
			readChan <- watchLineReaded(sc.Text())
		}
		if err := sc.Err(); err != nil && err != io.EOF {
			readChan <- watchErr(err)
		}
		readChan <- watchEof{}
	}()
	return readChan
}

func (r *watchReader) waitForLine() tea.Cmd {
	log("wait\n")
	return func() tea.Msg {
		s, ok := <-r.readChan
		if !ok {
			return watchChanClosed{}
		}
		return s
	}
}
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.wr.waitForLine(),
	)
}

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmds []tea.Cmd
//
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		cmdsWindow := m.updateWindowSize(msg)
// 		if cmdsWindow != nil {
// 			cmds = append(cmds, cmdsWindow...)
// 		}
// 	}
// 	var cmd tea.Cmd
// 	if m.viewport.IsMoveMsg(msg) != appendableviewport.NotMove {
// 		m.viewport, cmd = m.viewport.Update(msg)
// 		cmds = append(cmds, cmd)
// 		m.freezed = !m.freezed
// 	}
// 	return m, tea.Batch(cmds...)
// }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		log("key: " + msg.String() + "\n")
		switch msg.String() {
		// case "esc":
		// 	if m.table.Focused() {
		// 		m.table.Blur()
		// 	} else {
		// 		m.table.Focus()
		// 	}
		case "q", "ctrl+c":
			log("quit\n")
			return m, tea.Quit
		}
	case watchLineReaded:
		log(fmt.Sprintf("%v, %v\n", m.count+1, string(msg)))
		m.count++
		newRows := m.table.Rows()
		count := strconv.Itoa(m.count)
		newRows = append(newRows, table.Row{count, string(msg)})
		m.table.SetRows(newRows)
		return m, m.wr.waitForLine()
	case watchChanClosed:
		log("Chan closed")
	case watchEof:
		log("EOF")
	case watchErr:
		panic(msg)
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	log("render\n")
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}

func main() {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	logFile, err = tea.LogToFile("log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("Try piping in some text.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	sc := bufio.NewScanner(reader)
	sc.Split(bufio.ScanLines)

	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "msg", Width: 70},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(make([]table.Row, 0, 500)),
		table.WithFocused(true),
		table.WithHeight(7),
	)

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
	t.SetStyles(s)

	c := watchForLine(sc)

	wr := &watchReader{readChan: c}
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
