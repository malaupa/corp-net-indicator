package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	IsVerbose = false
	std       = log.New(os.Stderr, "", log.LstdFlags)
)

func Setup(prefix string, verbose bool) {
	IsVerbose = verbose
	std.SetPrefix(prefix + " ")
	if verbose {
		std.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		std.SetFlags(log.LstdFlags)
	}
}

func Log(v ...any) {
	std.Output(2, fmt.Sprintln(v...))
}

func Logf(format string, v ...any) {
	std.Output(2, fmt.Sprintf(format, v...))
}

func Verbose(v ...any) {
	if IsVerbose {
		std.Output(2, fmt.Sprintln(v...))
	}
}

func Verbosef(format string, v ...any) {
	if IsVerbose {
		std.Output(2, fmt.Sprintf(format, v...))
	}
}
