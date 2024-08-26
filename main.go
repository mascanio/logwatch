package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mascanio/logwatch/appendableViewport"
	// "github.com/mistakenelf/teacup/statusbar"

	"github.com/charmbracelet/bubbles/table"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type model struct {
	ready    bool
	viewport appendableviewport.Model
	table    table.Model
	wr       watchReader
	finished bool
	freezed  bool
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "c":
			m.freezed = false
		}
	case tea.WindowSizeMsg:
		cmdsWindow := m.updateWindowSize(msg)
		if cmdsWindow != nil {
			cmds = append(cmds, cmdsWindow...)
		}
	case watchLineReaded:
		m.viewport.Append(string(msg))
		cmds = append(cmds, m.wr.waitForLine())
		if !m.freezed {
			m.viewport.GotoBottom()
		}
		// else {
		// 	m.viewport.LineUp(1)
		// }
	case watchChanClosed:
		break
	case watchEof:
		// Wait untill chan is closed above
		m.finished = true
	case watchErr:
		panic(msg)
	}
	var cmd tea.Cmd
	if m.viewport.IsMoveMsg(msg) != appendableviewport.NotMove {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		m.freezed = !m.freezed
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	var s string
	if m.finished {
		s = "EOF"
	} else {
		s = "Running..."
	}
	title := titleStyle.Render(s)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("Try piping in some text.")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	sc := bufio.NewScanner(reader)
	sc.Split(bufio.ScanLines)

	p := tea.NewProgram(
		model{wr: watchReader{sc: sc, readChan: make(chan string)}},
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		tea.WithInputTTY(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

func (m *model) updateWindowSize(msg tea.WindowSizeMsg) []tea.Cmd {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	if !m.ready {
		// Since this program is using the full size of the viewport we
		// need to wait until we've received the window dimensions before
		// we can initialize the viewport. The initial dimensions come in
		// quickly, though asynchronously, which is why we wait for them
		// here.
		m.viewport = appendableviewport.New(msg.Width, msg.Height-verticalMarginHeight)
		m.viewport.YPosition = headerHeight
		m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
		m.viewport.SetContent(make([]string, 0))
		m.ready = true

		// This is only necessary for high performance rendering, which in
		// most cases you won't need.
		//
		// Render the viewport one line below the header.
		m.viewport.YPosition = headerHeight + 1
	} else {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
	}

	if useHighPerformanceRenderer {
		// Render (or re-render) the whole viewport. Necessary both to
		// initialize the viewport and when the window is resized.
		//
		// This is needed for high-performance rendering only.
		return []tea.Cmd{appendableviewport.Sync(m.viewport)}
	}
	return nil
}
