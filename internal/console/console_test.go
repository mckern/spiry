package console_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"

	// import console to test it
	"github.com/mckern/spiry/internal/console"
)

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
	{console.Debug, "electric kettle", "DEBUG"},
	{console.Warn, "electric can opener", "WARN"},
}

func TestRegularConsoleOutputs(t *testing.T) {
	for _, tcb := range consoleOutputTests {
		// We expect to see no output at all on stdin or stdout for
		// these functions without debugging enabled so we explicitly
		// disable debugging regardless of what we inherited from
		// our running environment
		disableDebugging(func() {
			out := captureOutput(func() { tcb.name(tcb.msg) })
			assert.Empty(t, out, "should contain no output from stdin or stderr")
		})
	}
}

func TestDebugConsoleOutputs(t *testing.T) {
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
	}
}

func TestErrorConsoleOutput(t *testing.T) {
	msg := "electric slicing knife"
	logLevel := "ERROR"

	// define some arbitrary message, in this case the execution
	// time of this particular test when it's run
	d := time.Now().Format("3:04PM")

	// define a semi-anonymous function, which will be appended
	// to a call to console.Error()
	f := func() { console.Error(fmt.Errorf("called appended func at %s", d)) }

	// call console.Error with the function appended, and then
	// check that the function actually ran by validating that
	// the output contains the timestamp (d) we expected to see
	stderr := captureStderr(func() { console.Error(errors.New(msg), f) })

	assert.Contains(t, stderr, d,
		fmt.Sprintf("should contain the date '%s' in output", d))
	assert.Contains(t, stderr, logLevel,
		fmt.Sprintf("should contain the log level '%s' in output", logLevel))
	assert.Contains(t, stderr, msg,
		"should contain the message '%s' in output", msg)
}

func TestFatalConsoleOutput(t *testing.T) {
	msg := "electric martini mixer"
	logLevel := "ERROR"

	// this is the test program that will be compiled and
	// run if the env. var. TEST_SUBSHELL is not set
	if os.Getenv("TEST_SUBSHELL") == "true" {
		console.Fatal(errors.New(msg))
		return
	}

	// invoke this test in a subshell, passing an environment
	// variable indicating that the test should run the stubbed
	// code above instead of evaluating the assertions below
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalConsoleOutput", "1>/dev/null")
	cmd.Env = append(os.Environ(), "TEST_SUBSHELL=true")
	stderr, exitCode := cmd.CombinedOutput()

	assert.Contains(t, string(stderr), logLevel,
		fmt.Sprintf("should contain the log level '%s' in output", logLevel))
	assert.Error(t, exitCode,
		"should cause program to exit with an error code")
	assert.Contains(t, string(stderr), msg,
		"should contain the message '%s' in output", msg)
}
