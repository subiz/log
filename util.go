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
//   skip: The number of stack frames to ignore from the beginning of the call stack (useful for skipping internal helper functions).
//
// Returns:
//   string: A formatted string representing the stack trace, with each frame in the format "file:line" and separated by " | ".
//   string: The name of the function at the top of the captured stack trace (after skipping).
//   string: A pipe-separated string of all function names in the call stack (after skipping).
func GetStack(skip int) (string, string, string) {
	stack := make([]uintptr, 10)
	var sb strings.Builder
	// skip one system stack, the this current stack line
	length := runtime.Callers(4+skip, stack[:])
	funcname := ""
	first := -1
	funcstack := ""
	for i := 0; i < length; i++ {
		pc := stack[i]
		// pc - 1 because the program counters we use are usually return addresses,
		// and we want to show the line that corresponds to the function call
		function := runtime.FuncForPC(pc - 1)
		// funcstack=/github.com/subiz/log.EServer | github.com/subiz/log_test.E | github.com/subiz/log_test.DDDDDD | github.com/subiz/log_test.CCCCCC | github.com/subiz/log_test.B | github.com/subiz/log_test.A | github.com/subiz/log_test.TestError | testing.tRunner | runtime.goexit
		funcstack += function.Name() + " | "
		if i == 0 {
			funcname = function.Name()
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

		if first == -1 {
			first = i
		} else {
			sb.WriteString(" | ")
		}
		sb.WriteString(file + ":" + strconv.Itoa(line))
	}
	return sb.String(), funcname, funcstack
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
