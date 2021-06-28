package log

import (
	"runtime"
	"strconv"
	"strings"
)

// find outside caller
func getCaller() string {
	// fast lookup
	_, currentFile, currentLine, _ := runtime.Caller(4)
	return chopPath(currentFile) + " " + strconv.Itoa(currentLine)
}

var defaultPaths = []string{
	"/src/git.subiz.net/",
	"/src/github.com/subiz/",
}

func chopPath(path string) string {
	for _, p := range defaultPaths {
		i := strings.LastIndex(path, p)
		if i >= 0 {
			return path[i+len(p):]
		}
	}
	return path
}

// isSystemPath tells whether a file is in system golang packages
func isSystemPath(path string) bool {
	if strings.Contains(path, "/github.com/subiz/errors/") {
		return true
	}
	return strings.HasPrefix(path, "/usr/local/go/src")
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

// getStack returns 10 closest stacktrace, included file paths and line numbers
// it will ignore all system path, path which is vendor is striped to /vendor/
// skip: number of stack ignored
func getStack(skip int) string {
	stack := make([]uintptr, 10)
	var sb strings.Builder
	// skip one system stack, the this current stack line
	length := runtime.Callers(2+skip, stack[:])
	for i := 0; i < length; i++ {
		pc := stack[i]
		// pc - 1 because the program counters we use are usually return addresses,
		// and we want to show the line that corresponds to the function call
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc - 1)
		// dont report system path
		if isSystemPath(file) {
			continue
		}

		file = trimToPrefix(file, "/vendor/")

		// trim out common provider since most of go projects are hosted
		// in single host, there is no need to include them in the call stack
		// remove them help keeping the call stack smaller, navigatiing easier
		if !strings.HasPrefix(file, "/vendor") {
			file = trimOutPrefix(file, "/git.subiz.net/")
			file = trimOutPrefix(file, "/github.com/")
			file = trimOutPrefix(file, "/gitlab.com/")
			file = trimOutPrefix(file, "/gopkg.in/")
		}

		sb.WriteString(file)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(line))
		sb.WriteString("$")
	}
	return strings.TrimSpace(sb.String())
}
