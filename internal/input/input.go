package input

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	tty "github.com/mattn/go-tty"
)

type Input interface {
	io.ReadCloser
}

type InputExec struct {
	cmd     *exec.Cmd
	outPipe io.ReadCloser
}

func NewExec(c []string) InputExec {
	cmd := exec.Command(c[0], c[1:]...)
	outPipe, errExec := cmd.StdoutPipe()
	if errExec != nil {
		panic(errExec)
	}
	if err := cmd.Start(); err != nil {
		outPipe.Close()
		panic(err)
	}
	return InputExec{cmd, outPipe}
}

func (i *InputExec) Read(p []byte) (int, error) { return i.outPipe.Read(p) }
func (i *InputExec) Close() error {
	if i.cmd.ProcessState.Exited() {
		return nil
	}
	return i.cmd.Process.Kill()
}

type InputBasicPipe struct {
	io.ReadCloser
}

func NewBasicPipe() (InputBasicPipe, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return InputBasicPipe{}, err
	}
	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		return InputBasicPipe{}, fmt.Errorf("Try piping in some text.")
	}
	return InputBasicPipe{os.Stdin}, nil
}

type InputTTY struct {
	t *tty.TTY
}

func NewTTY() (InputTTY, error) {
	tty, err := tty.Open()
	if err != nil {
		return InputTTY{}, err
	}
	return InputTTY{tty}, nil
}

func (i *InputTTY) Read(p []byte) (int, error) { return i.t.Input().Read(p) }
func (i *InputTTY) Close() error               { return i.t.Close() }
