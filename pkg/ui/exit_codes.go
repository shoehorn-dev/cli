package ui

import (
	"fmt"
	"os"
	"strings"
)

// Exit codes for the CLI
const (
	ExitSuccess      = 0 // Command executed successfully
	ExitError        = 1 // Generic error
	ExitAuthRequired = 2 // Authentication required
	ExitNotFound     = 3 // Resource not found
	ExitValidation   = 4 // Validation error
	ExitTimeout      = 5 // Operation timeout
	ExitCancelled    = 6 // User cancelled operation
)

// Exit terminates the program with the given exit code
func Exit(code int) {
	os.Exit(code)
}

// ExitWithError prints the error and exits with appropriate code
func ExitWithError(err error) {
	if err == nil {
		Exit(ExitSuccess)
		return
	}

	// Print error to stderr
	RenderError(err)

	// Determine exit code based on error message
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not authenticated"),
		strings.Contains(errMsg, "authentication required"),
		strings.Contains(errMsg, "401"):
		Exit(ExitAuthRequired)

	case strings.Contains(errMsg, "not found"),
		strings.Contains(errMsg, "404"):
		Exit(ExitNotFound)

	case strings.Contains(errMsg, "validation"),
		strings.Contains(errMsg, "invalid"),
		strings.Contains(errMsg, "400"):
		Exit(ExitValidation)

	case strings.Contains(errMsg, "timeout"),
		strings.Contains(errMsg, "deadline exceeded"):
		Exit(ExitTimeout)

	case strings.Contains(errMsg, "cancelled"),
		strings.Contains(errMsg, "canceled"):
		Exit(ExitCancelled)

	default:
		Exit(ExitError)
	}
}

// ExitWithMessage prints a message and exits with the given code
func ExitWithMessage(code int, message string, args ...interface{}) {
	if code == ExitSuccess {
		fmt.Fprintf(os.Stdout, message+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, message+"\n", args...)
	}
	Exit(code)
}
