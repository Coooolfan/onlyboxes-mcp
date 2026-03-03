package logging

import (
	"fmt"
	"log"
	"strings"
)

const (
	colorReset = "\033[0m"
	colorInfo  = "\033[1;36m"
	colorWarn  = "\033[1;33m"
	colorError = "\033[1;31m"
	colorFatal = "\033[1;35m"
)

var singleLineReplacer = strings.NewReplacer(
	"\r\n", "\\n",
	"\n", "\\n",
	"\r", "\\r",
)

func formatSingleLine(format string, args ...any) string {
	return singleLineReplacer.Replace(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	log.Printf("%s[INFO]%s %s", colorInfo, colorReset, formatSingleLine(format, args...))
}

func Warnf(format string, args ...any) {
	log.Printf("%s[WARN]%s %s", colorWarn, colorReset, formatSingleLine(format, args...))
}

func Errorf(format string, args ...any) {
	log.Printf("%s[ERROR]%s %s", colorError, colorReset, formatSingleLine(format, args...))
}

func Fatalf(format string, args ...any) {
	log.Fatalf("%s[FATAL]%s %s", colorFatal, colorReset, formatSingleLine(format, args...))
}
