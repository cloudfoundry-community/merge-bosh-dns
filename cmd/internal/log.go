package internal

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Log(format string, args ...interface{}) {
	fLog(os.Stdout, format, args)
}

func LogErr(format string, args ...interface{}) {
	fLog(os.Stderr, "ERROR: "+format, args)
}

func fLog(out io.Writer, format string, args ...interface{}) {
	now := time.Now()
	fmt.Fprintf(out, now.Format(time.RFC3339)+": "+format+"\n", args...)
}
