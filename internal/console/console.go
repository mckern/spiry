package console

import (
	"fmt"
	"os"
	"strings"
)

var DebugVariable = "DEBUG"

func Debug(msg string) {
	_, defined := os.LookupEnv(DebugVariable)
	if defined {
		fmt.Fprintf(os.Stderr, "DEBUG: %s\n", strings.TrimSpace(msg))
	}
}

func Warn(msg string) {
	_, defined := os.LookupEnv(DebugVariable)
	if defined {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", strings.TrimSpace(msg))
	}
}

func Error(err error, fs ...func()) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())

		// call any funcs passed along with the error
		for _, f := range fs {
			f()
		}
	}
}

func Fatal(err error, fs ...func()) {
	fs = append(fs, func() { os.Exit(1) })
	Error(err, fs...)
}
