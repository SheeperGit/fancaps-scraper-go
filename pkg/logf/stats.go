package logf

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

/* Log statistics */
type LogStats struct {
	stats map[LogSeverity]int
	mu    sync.RWMutex
}

/* Keeps track of log statistics. */
var logStats = &LogStats{
	stats: make(map[LogSeverity]int, len(SeverityName)),
}

/* Increments the log severity `logSev` statistic  of log statistics `ls`. */
func Increment(logSev LogSeverity) {
	logStats.mu.Lock()
	defer logStats.mu.Unlock()

	logStats.stats[logSev]++
}

/*
Returns the log statistics.

Includes information about the amount of logs for each log severity.
*/
func Stats() map[LogSeverity]int {
	logStats.mu.RLock()
	defer logStats.mu.RUnlock()

	copy := make(map[LogSeverity]int, len(SeverityName))
	for sev := range SeverityName {
		copy[sev] = logStats.stats[sev]
	}

	return copy
}

/*
Prints log statistics.
If there are no statistics, this function prints a message indicating
the operation has completed successfully.
*/
func PrintStats() {
	fmt.Printf("\n\n")

	if Logfile != "" {
		fmt.Fprintln(os.Stderr, "Log Summary:")
		for stat, amt := range Stats() {
			var style lipgloss.Style
			switch amt {
			case 0:
				style = ui.SuccessStyle
			default:
				style = ui.ErrStyle
			}
			fmt.Fprintf(os.Stderr, style.Render("\t%s: %d")+"\n", stat, amt)
		}
		fmt.Fprintln(os.Stderr)

		fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("Logs found which may require your attention."))
		fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("Check logfile for further details:"))
		fmt.Fprintf(os.Stderr, "\t%s\n", Logfile)
	} else {
		fmt.Println(ui.SuccessStyle.Render("Operation completed successfully!"))
	}
}
