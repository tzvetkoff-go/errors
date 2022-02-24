// Package errors provides errors with stacktraces.
package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

// FormatType ...
type FormatType int

// FormatTypes ...
const (
	FormatFull FormatType = 1 << iota
	FormatShort
)

// DefaultFormat ...
var DefaultFormat = FormatFull

// StripPath ...
var StripPath = func(p string) string {
	dirs := filepath.SplitList(os.Getenv("GOPATH"))
	sort.SliceStable(dirs, func(i int, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, dir := range dirs {
		src := filepath.Join(dir, "src") + string(os.PathSeparator)
		if strings.HasPrefix(p, src+src) {
			return p[len(src)+1:]
		}
	}

	return p
}

// Error ...
type Error struct {
	Message  string
	Cause    error
	File     string
	Function string
	Line     int
}

// New ...
func New(format string, args ...interface{}) error {
	return Create(nil, format, args...)
}

// Propagate ...
func Propagate(cause error, format string, args ...interface{}) error {
	if cause == nil {
		return nil
	}

	return Create(cause, format, args...)
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

// Create ...
func Create(cause error, format string, args ...interface{}) error {
	err := &Error{
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}

	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return err
	}

	err.File = StripPath(file)
	err.Line = line

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return err
	}

	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	funcName := fn.Name()
	funcName = funcName[strings.LastIndex(funcName, "/")+1:]
	funcName = funcName[strings.Index(funcName, ".")+1:]
	funcName = strings.Replace(funcName, "(", "", 1)
	funcName = strings.Replace(funcName, "*", "", 1)
	funcName = strings.Replace(funcName, ")", "", 1)
	err.Function = funcName

	return err
}

// Unwrap ...
func Unwrap(err error) error {
	if err, ok := err.(*Error); ok {
		if err.Cause != nil {
			return err.Cause
		}
	}

	if unwrappable, ok := err.(interface{ Unwrap() error }); ok {
		return unwrappable.Unwrap()
	}

	return nil
}

// Is ...
func Is(err, target error) bool {
	if target == nil {
		return err == nil
	}

	comparable := reflect.TypeOf(target).Comparable()
	for {
		if comparable && err == target {
			return true
		}

		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}

		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

// As ...
func As(err error, target interface{}) bool {
	if target == nil {
		return false
	}

	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}
	targetType := typ.Elem()

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if targetType.Kind() != reflect.Interface && !targetType.Implements(errorType) {
		return false
	}

	for err != nil {
		if reflect.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}

		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}

		err = Unwrap(err)
	}

	return false
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
		text = formatShort(e)
	} else if DefaultFormat == FormatFull {
		text = formatFull(e)
	} else {
		text = formatShort(e)
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

// formatShort ...
func formatShort(e *Error) string {
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
