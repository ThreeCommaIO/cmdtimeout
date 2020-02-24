package cmdtimeout

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
)

func TestCmdTimeoutLeak(t *testing.T) {
	defer leaktest.Check(t)()

	timeout := 1 * time.Second
	args := "100"
	cmd := exec.Command("sleep", strings.Split(args, " ")...)
	ct := New(cmd, timeout)
	if err := ct.Start(); err != ErrTimeoutHit {
		t.Error(err)
	}
}
