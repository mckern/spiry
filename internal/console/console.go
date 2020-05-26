package console

import (
	"fmt"
	"os"
	"strings"
)

func Debug(msg string) {
	value, defined := os.LookupEnv("SPIRY_DEBUG")
	if defined && value != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: %s\n", strings.Trim(msg, "\n"))
	}
}

func Warn(msg string) {
	value, defined := os.LookupEnv("SPIRY_DEBUG")
	if defined && value != "" {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", strings.Trim(msg, "\n"))
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
	if err != nil {
		Error(err, fs...)
	}
}
