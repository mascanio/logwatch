package main

import (
	"bufio"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mascanio/logwatch/internal/input"
	"github.com/mascanio/logwatch/internal/models"
)

var logFile *os.File

func log(s string) {
	if _, err := logFile.WriteString(s); err != nil {
		panic(err)
	}
}

func main() {
	var err error
	logFile, err = tea.LogToFile("log", "debug")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	inputStream, err := input.NewBasicPipe()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(inputStream)
	sc := bufio.NewScanner(reader)
	sc.Split(bufio.ScanLines)

	p := tea.NewProgram(
		models.NewGlobal(sc),
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		tea.WithInputTTY(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
