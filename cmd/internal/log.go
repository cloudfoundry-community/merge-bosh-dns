package internal

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Log(format string, args ...interface{}) {
	if args != nil {
		fLog(os.Stdout, format, args)
	} else {
		fLog(os.Stdout, format)
	}
}

func LogErr(format string, args ...interface{}) {
	if args != nil {
		fLog(os.Stderr, "ERROR: "+format, args)
	} else {
		fLog(os.Stderr, "ERROR: "+format)
	}
}

func fLog(out io.Writer, format string, args ...interface{}) {
	now := time.Now()
	fmt.Fprintf(out, now.Format(time.RFC3339)+": "+format+"\n", args...)
}
