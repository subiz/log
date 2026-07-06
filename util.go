package log

import (
	"runtime"
	"strconv"
	"strings"
)

// find outside caller
func getCaller() string {
	// fast lookup
	_, currentFile, currentLine, _ := runtime.Caller(3)
	return currentFile + ":" + strconv.Itoa(currentLine)
}

// trimOutPrefix removes all the characters before AND the prefix
// its return the original string if not found prefix in str
func trimOutPrefix(str, prefix string) string {
	i := strings.Index(str, prefix)
	if i < 0 {
		return str
	}
	return str[i+len(prefix):]
}

// trimToPrefix removes all the characters before the prefix
// its return the original string if not found prefix in str
func trimToPrefix(str, prefix string) string {
	i := strings.Index(str, prefix)
	if i < 0 {
		return str
	}
	return str[i:]
}

// GetStack returns a formatted stack trace, the name of the function that initiated the stack trace,
// and a pipe-separated string of function names in the call stack.
// It captures up to 10 closest stack frames, including file paths and line numbers.
// System paths are ignored, and paths within the 'vendor' directory are truncated to '/vendor/'.
//
// Parameters:
//
//	skip: The number of stack frames to ignore from the beginning of the call stack (useful for skipping internal helper functions).
//
// Returns:
//
//	string: A formatted string representing the stack trace, with each frame in the format "file:line" and separated by " | ".
//	string: The name of the function at the top of the captured stack trace (after skipping).
//	string: A pipe-separated string of all function names in the call stack (after skipping).
func GetStack(skip int) (string, string, string) {
	// stack-allocated array avoids a heap allocation for the common case
	var pcs [10]uintptr
	// skip one system stack, the this current stack line
	length := runtime.Callers(4+skip, pcs[:])

	var sb strings.Builder     // the "file:line" stack
	var funcsb strings.Builder // the pipe-separated function names
	var numbuf [20]byte        // scratch for the line number, avoids strconv.Itoa allocs
	funcname := ""
	first := true
	for i := 0; i < length; i++ {
		pc := pcs[i]
		// pc - 1 because the program counters we use are usually return addresses,
		// and we want to show the line that corresponds to the function call
		function := runtime.FuncForPC(pc - 1)
		name := function.Name()

		funcsb.WriteString(name)
		funcsb.WriteString(" | ")
		if i == 0 {
			funcname = name
		}

		// dont report the Go runtime frames (e.g. runtime.main, runtime.goexit);
		// their file path is the Go SDK, which varies by install location, so
		// filter by function package instead of by path. checked before FileLine
		// (comparatively expensive) so skipped frames don't pay for it.
		if strings.HasPrefix(name, "runtime.") {
			continue
		}

		file, line := function.FileLine(pc - 1)
		file = trimToPrefix(file, "/vendor/")
		// dont report system path
		if isIgnorePath(file) {
			continue
		}

		// trim out common provider since most of go projects are hosted
		// in single host, there is no need to include them in the call stack
		// remove them help keeping the call stack smaller, navigatiing easier
		if !strings.HasPrefix(file, "/vendor") {
			file = trimOutPrefix(file, "/github.com/")
			file = trimOutPrefix(file, "/gitlab.com/")
			file = trimOutPrefix(file, "/gopkg.in/")
		}

		if first {
			first = false
		} else {
			sb.WriteString(" | ")
		}
		sb.WriteString(hostname)
		if !strings.HasPrefix(file, "/") {
			sb.WriteByte('/')
		}
		sb.WriteString(file)
		sb.WriteByte(':')
		sb.Write(strconv.AppendInt(numbuf[:0], int64(line), 10))
	}
	return sb.String(), funcname, funcsb.String()
}

func Stack() string {
	stack, _, _ := GetStack(-1)
	return stack
}

// isIgnorePath indicates whether a path is just noise, that excluding the path does not
// affect error context
func isIgnorePath(path string) bool {
	if strings.HasPrefix(path, "/usr/local/go/src") {
		return true
	}

	if strings.HasPrefix(path, "/vendor/google.golang.org/") {
		return true
	}

	if strings.HasPrefix(path, "/vendor/github.com/gin-gonic") {
		return true
	}
	return false
}
