package logging

import (
	"fmt"
	"log"
)

const (
	colorReset = "\033[0m"
	colorInfo  = "\033[1;36m"
	colorWarn  = "\033[1;33m"
	colorError = "\033[1;31m"
	colorFatal = "\033[1;35m"
)

func Infof(format string, args ...any) {
	log.Printf("%s[INFO]%s %s", colorInfo, colorReset, fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	log.Printf("%s[WARN]%s %s", colorWarn, colorReset, fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	log.Printf("%s[ERROR]%s %s", colorError, colorReset, fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...any) {
	log.Fatalf("%s[FATAL]%s %s", colorFatal, colorReset, fmt.Sprintf(format, args...))
}
