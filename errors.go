// Package errors provides errors with stacktraces.
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// FormatTypes ...
const (
	FormatFull = iota
	FormatBrief
)

// DefaultFormat ...
var DefaultFormat = FormatFull

// Error ...
type Error struct {
	Message  string
	Cause    error
	File     string
	Function string
	Line     int
}

// Cause ...
func Cause(err error) error {
	if err, ok := err.(*Error); ok {
		if err.Cause != nil {
			return Cause(err.Cause)
		}
	}

	return err
}

// New ...
func New(format string, args ...interface{}) error {
	return create(nil, format, args...)
}

// Propagate ...
func Propagate(cause error, format string, args ...interface{}) error {
	if cause == nil {
		return nil
	}

	return create(cause, format, args...)
}

// create ...
func create(cause error, format string, args ...interface{}) error {
	err := &Error{
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}

	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return err
	}

	err.File = file
	err.Line = line

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return err
	}
	err.Function = shortFuncName(fn)

	return err
}

// shortFuncName ...
func shortFuncName(f *runtime.Func) string {
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}

// Error ...
func (e *Error) Error() string {
	return fmt.Sprint(e)
}

// Format ...
func (e *Error) Format(f fmt.State, c rune) {
	var text string
	if f.Flag('+') && !f.Flag('#') && c == 's' { // "%+s"
		text = formatFull(e)
	} else if f.Flag('#') && !f.Flag('+') && c == 's' { // "%#s"
		text = formatBrief(e)
	} else if DefaultFormat == FormatFull {
		text = formatFull(e)
	} else {
		text = formatBrief(e)
	}

	formatString := "%"
	for _, flag := range "-+# 0" {
		if f.Flag(int(flag)) {
			formatString += string(flag)
		}
	}
	if width, has := f.Width(); has {
		formatString += fmt.Sprint(width)
	}
	if precision, has := f.Precision(); has {
		formatString += "."
		formatString += fmt.Sprint(precision)
	}
	formatString += string(c)
	fmt.Fprintf(f, formatString, text)
}

// formatFull ...
func formatFull(e *Error) string {
	s := ""

	newline := func() {
		if s != "" && !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
	}

	for curr, ok := e, true; ok; curr, ok = curr.Cause.(*Error) {
		s += curr.Message

		if curr.File != "" {
			newline()

			if curr.Function == "" {
				s += fmt.Sprintf(" --- at %v:%v ---", curr.File, curr.Line)
			} else {
				s += fmt.Sprintf(" --- at %v:%v (%v) ---", curr.File, curr.Line, curr.Function)
			}
		}

		if curr.Cause != nil {
			newline()
			if cause, ok := curr.Cause.(*Error); !ok {
				s += "Caused by: "
				s += curr.Cause.Error()
			} else if cause.Message != "" {
				s += "Caused by: "
			}
		}
	}

	return s
}

// formatBrief ...
func formatBrief(e *Error) string {
	s := ""

	concat := func(msg string) {
		if s != "" && msg != "" {
			s += ": "
		}
		s += msg
	}

	curr := e
	for {
		concat(curr.Message)
		if cause, ok := curr.Cause.(*Error); ok {
			curr = cause
		} else {
			break
		}
	}
	if curr.Cause != nil {
		concat(curr.Cause.Error())
	}

	return s
}
