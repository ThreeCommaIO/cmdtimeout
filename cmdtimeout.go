package cmdtimeout

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"
)

// errors
var (
	ErrTimeoutHit = errors.New("timeout was hit")
)

// CmdTimeout handles the options for configuring timers for commands
type CmdTimeout struct {
	cmd      *exec.Cmd
	duration time.Duration
	done     chan error
	data     chan string
	stdout   io.ReadCloser
	stderr   io.ReadCloser
}

// New creates a new command timer
func New(cmd *exec.Cmd, duration time.Duration) *CmdTimeout {
	done := make(chan error, 1)
	data := make(chan string, 1)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	return &CmdTimeout{cmd, duration, done, data, stdout, stderr}
}

// Start handles running the command and waiting for a timeout
func (c *CmdTimeout) Start() error {
	t := time.NewTimer(c.duration)
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(c.stdout)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			line := scanner.Text()
			c.data <- line
		}
	}()

	go func() {
		defer wg.Done()

		scanner := bufio.NewScanner(c.stdout)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			line := scanner.Text()
			c.data <- line
		}
	}()

	// start the command
	if err := c.cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer wg.Done()
		c.done <- c.cmd.Wait()
	}()

	go func() {
		wg.Wait()
	}()

readLine:
	for {
		select {
		case <-c.data:
			if !t.Stop() {
				<-t.C
			}
			t.Reset(c.duration)
			goto readLine
		case <-t.C:
			c.cmd.Process.Kill()
			return ErrTimeoutHit
		case err := <-c.done:
			if err != nil {
				return err
			}
			return nil
		}
	}

}
