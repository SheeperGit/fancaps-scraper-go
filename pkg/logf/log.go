package logf

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
)

/* Enum for log severity. */
type LogSeverity int

const (
	LOG_ERROR   LogSeverity = iota // Critical log severity.
	LOG_WARNING                    // Non-critical log severity.
)

/* Convert a log severity enumeration to its corresponding string representation. */
func (logsev LogSeverity) String() string {
	return SeverityName[logsev]
}

var SeverityName = map[LogSeverity]string{
	LOG_ERROR:   "ERROR",
	LOG_WARNING: "WARNING",
}

var (
	setOnce        sync.Once // Initializes certain logging variables.
	Logfile        string    // Path to log file. Contains logs of varying severity. Non-empty, if something unexpected happened.
	maxSeverityLen int       // Maximum length string of a log severity.
)

/*
Appends errors to a log file, as defined by its severity `severity`, format `format` and its
arguments `args`. Errors are timestamped with nanosecond precision.
*/
func LogErrorf(logSev LogSeverity, format string, args ...any) {
	flags := cli.Flags()
	if flags.NoLog {
		return
	}

	setOnce.Do(func() {
		fileTimestamp := time.Now().Format("2006-01-02_15-04-05.000000000") // Nanosecond precision.
		LogDir := flags.OutputDir
		Logfile = filepath.Join(LogDir, fmt.Sprintf("fsg_errors_%s.txt", fileTimestamp))

		maxSeverityLen := 0
		for _, name := range SeverityName {
			if len(name) > maxSeverityLen {
				maxSeverityLen = len(name)
			}
		}
	})

	f, err := os.OpenFile(Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open logfile: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	/* Write log to file. */
	errTimestamp := time.Now().Format("2006-01-02 15:04:05.000000000")        // Nanosecond precision.
	sev := fmt.Sprintf("%-*s", maxSeverityLen+2, fmt.Sprintf("[%s]", logSev)) // Left-align severity error text.
	errLine := fmt.Sprintf("%s (%s) %s\n", sev, errTimestamp, fmt.Sprintf(format, args...))
	f.WriteString(errLine)

	/* Update statistics. */
	Increment(logSev)
}
