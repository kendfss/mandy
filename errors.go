package mandy

import (
	"errors"
	"strconv"
)

// These constants cause Command.Parse to behave as described if the parse fails.
const (
	ContinueOnError ErrorPolicy = iota // Return a descriptive error.
	ExitOnError                        // Call os.Exit(2) or for -h/-help Exit(0).
	PanicOnError                       // Call panic with a descriptive error.
	LogOnError                         // Write a descriptive error to os.Stderr.
)

var (
	// ErrHelp is the error returned if the -help or -h flag is invoked
	// but no such flag is defined.
	ErrHelp = errors.New("mandy: help requested")

	// The error returned when a command that has no main function is Executed
	ErrNilMain = errors.New("mandy: attempted to Execute a command with no Main function")

	// errParse is returned by Set if a flag's value fails to parse, such as with an invalid integer for Int.
	// It then gets wrapped through failf to provide more information.
	errParse = errors.New("parse error")

	// errRange is returned by Set if a flag's value is out of range.
	// It then gets wrapped through failf to provide more information.
	errRange = errors.New("value out of range")
)

// ErrorPolicy defines how Command.Parse behaves if the parse fails.
type ErrorPolicy uint8

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return err
}
