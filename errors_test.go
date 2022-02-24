package errors_test

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/tzvetkoff-go/errors"
)

func TestPropagate(t *testing.T) {
	f1 := func() error {
		return errors.New("error from f3")
	}
	f2 := func() error {
		return errors.Propagate(f1(), "error from f2")
	}
	f3 := func() error {
		return errors.Propagate(f2(), "error from f1")
	}

	err := f3()
	expected := `error from f1
 --- at errors_test.go:23 (TestPropagate.func3) ---
Caused by: error from f2
 --- at errors_test.go:20 (TestPropagate.func2) ---
Caused by: error from f3
 --- at errors_test.go:17 (TestPropagate.func1) ---`

	if fmt.Sprint(err) != expected {
		t.Errorf("TestStacktrace: got:\n[\n%s\n], expected:\n[\n%s\n]", err, expected)
	}
}

func TestUnwrap(t *testing.T) {
	error1 := errors.New("error1")
	error2 := errors.New("error2")
	error3 := errors.Propagate(error2, "error3")

	testCases := []struct {
		where    string
		err      error
		expected error
	}{
		{here(), error1, nil},
		{here(), error2, nil},
		{here(), error3, error2},
	}

	for _, tc := range testCases {
		got := errors.Unwrap(tc.err)
		if got != tc.expected {
			t.Errorf("%s: got %q, expected %q", tc.where, tc.err, tc.expected)
		}
	}
}

func TestIs(t *testing.T) {
	error1 := errors.New("error1")
	error2 := errors.Propagate(error1, "error2")
	error3 := errors.Propagate(error2, "error3")

	testCases := []struct {
		where    string
		err      error
		target   error
		expected bool
	}{
		{here(), nil, nil, true},
		{here(), error1, nil, false},
		{here(), error1, error1, true},
		{here(), error2, error1, true},
		{here(), error3, error1, true},
	}

	for _, tc := range testCases {
		got := errors.Is(tc.err, tc.target)
		if got != tc.expected {
			t.Errorf("%s: errors.Is(%q, %q), got %v, expected %v", tc.where, tc.err, tc.target, got, tc.expected)
		}
	}
}

func TestAs(t *testing.T) {
	error1 := &fs.PathError{Op: "readdir", Path: "error1", Err: errors.New("root-error")}
	error2 := errors.Propagate(error1, "error2")
	error3 := errors.New("error3")

	testCases := []struct {
		where    string
		err      error
		target   *fs.PathError
		expected bool
	}{
		{here(), nil, nil, false},
		{here(), error1, nil, true},
		{here(), error2, nil, true},
		{here(), error3, nil, false},
	}

	for _, tc := range testCases {
		got := errors.As(tc.err, &tc.target)
		if got != tc.expected {
			t.Errorf("%s: errors.As(%q, *fs.PathError), got %v, expected %v", tc.where, tc.err, got, tc.expected)
		}
	}

	var target *fs.PathError
	if !errors.As(error2, &target) {
		t.Errorf("errors.As(%q, *fs.PathError), got %v, expected %v", error2, false, true)
	}
}

func here() string {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%s:%d", path.Base(file), line)
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	root := path.Join(path.Dir(file), string(os.PathSeparator))

	origStripPath := errors.StripPath
	errors.StripPath = func(p string) string {
		if strings.HasPrefix(p, root) {
			return p[len(root)+1:]
		}

		return origStripPath(p)
	}
}
