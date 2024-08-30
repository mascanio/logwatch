package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mascanio/logwatch/internal/input"
	"github.com/mascanio/logwatch/internal/models/global"
)

func main() {
	// Setup logger
	logFile, err := tea.LogToFile("log", "debug")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	inputStream, err := input.NewBasicPipe()
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(inputStream)

	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "msg", Width: 70},
	}

	model := global.New(sc,
		global.WithTableColums(columns),
		global.WithScanner(sc),
	)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		tea.WithInputTTY(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
