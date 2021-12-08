package modbus

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

var traceEnabled = false

func init() {
	if os.Getenv("GO_MODBUS_TRACE") == "true" {
		traceEnabled = true
	}
}

func Trace(args ...interface{}) {
	if traceEnabled {
		log.Println(args...)
	}
}

func durationMs(ms int) time.Duration {
	return time.Millisecond * time.Duration(ms)
}

func unixMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

func formatErr(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s %s", msg, string(debug.Stack()))
}
