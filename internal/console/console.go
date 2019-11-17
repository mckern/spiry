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

func Fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}
