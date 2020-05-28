package console

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
)

// trim whitespace off the output,
// to get rid of that trailing newline
// func captureStdout(f func()) string {
// 	return strings.TrimSpace(capturer.CaptureStdout(f))
// }

func captureStderr(f func()) string {
	return strings.TrimSpace(capturer.CaptureStderr(f))
}

func captureOutput(f func()) string {
	return strings.TrimSpace(capturer.CaptureOutput(f))
}

func enableDebugging(f func()) {
	originalValue := os.Getenv("SPIRY_DEBUG")
	_ = os.Setenv("SPIRY_DEBUG", "true")
	f()
	os.Setenv("SPIRY_DEBUG", originalValue)
}

func disableDebugging(f func()) {
	originalValue := os.Getenv("SPIRY_DEBUG")
	os.Unsetenv("SPIRY_DEBUG")
	f()
	os.Setenv("SPIRY_DEBUG", originalValue)
}

var consoleOutputTests = []struct {
	name  func(s string) // name of function to run
	msg   string         // message to feed it
	level string         // log level to expect it to emit
}{
	{Debug, "electric kettle", "DEBUG"},
	{Warn, "electric can opener", "WARN"},
}

func TestConsoleOutputs(t *testing.T) {
	for _, tcb := range consoleOutputTests {
		// these functions won't print output unless debugging
		// is enabled; so we mangle the environment briefly to
		// ensure that output redirects appropriately.
		enableDebugging(func() {
			out := captureStderr(func() { tcb.name(tcb.msg) })

			assert.Contains(t, out, tcb.level, "should contain the log level in output")
			assert.Contains(t, out, tcb.msg,
				fmt.Sprintf(`should contain the message "%s" in output`, tcb.msg))
		})

		// We expect to see no output at all on stdin or stdout for
		// these functions without debugging enabled
		disableDebugging(func() {
			out := captureOutput(func() { tcb.name(tcb.msg) })
			assert.Empty(t, out, "should contain no output from stdin or stderr")
		})
	}
}
