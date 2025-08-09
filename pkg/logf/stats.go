package logf

import (
	"fmt"
	"os"
	"slices"
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

/* Prints log statistics. */
func PrintStats() {
	if Logfile != "" {
		fmt.Fprintln(os.Stderr, "\n\nLog Summary:")

		/* Get stats. */
		stats := Stats()
		sevStats := []LogSeverity{}
		for sev := range stats {
			sevStats = append(sevStats, sev)
		}
		slices.Sort(sevStats) // Sort severity stats by enum order.

		for _, sev := range sevStats {
			amt := stats[sev]

			var style lipgloss.Style
			switch amt {
			case 0:
				style = ui.SuccessStyle
			default:
				style = ui.ErrStyle
			}
			fmt.Fprintf(os.Stderr, "\t"+style.Render("%s: %d")+"\n", SeverityName[sev], amt)
		}

		fmt.Fprintf(os.Stderr, "\n"+
			ui.ErrStyle.Render("Logs found which may interest you.")+"\n"+
			ui.ErrStyle.Render("Check logfile for further details:")+"\n"+
			"\t%s\n", Logfile)
	}
}
