package logger

import "log"

var IsVerbose = false

func Setup(prefix string, verbose bool) {
	IsVerbose = verbose
	log.SetPrefix(prefix)
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
}

func Log(v any) {
	log.Println(v)
}

func Logf(format string, v ...any) {
	log.Printf(format, v...)
}

func Verbose(v any) {
	if IsVerbose {
		log.Println(v)
	}
}

func Verbosef(format string, v ...any) {
	if IsVerbose {
		log.Printf(format, v...)
	}
}
