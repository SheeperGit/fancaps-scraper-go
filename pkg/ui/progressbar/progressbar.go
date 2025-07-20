package progressbar

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

const (
	saucer           = "#" // Character indicating complete progress.
	saucerPadding    = "-" // Character indicating incomplete progress.
	progressWidth    = 52  // Amount of glyphs within the progress bar.
	defaultTermWidth = 80  // Default terminal width. (Fallback)
	titleSpacing     = 1   // Spacing applied to title names on the progress bar.
	episodeSpacing   = 3   // Spacing applied to episode names on the progress bar.
	ratioWidth       = 13  // Width taken by the ratio of completed/total images downloaded in the progress bar.
	percentageWidth  = 4   // Width taken by the percentage of completed/total images downloaded in the progress bar.
)

var (
	progressMu       sync.Mutex // Limits progress bar access to one thread.
	lastPrintedLines int        // Number of printed lines on the last call to `ShowProgress()`.
)

/* Displays progress bar(s) based on the state of the titles `titles`. */
func ShowProgress(titles []*types.Title) {
	progressMu.Lock()
	defer progressMu.Unlock()

	termWidth, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get terminal width. %v\nDefaulting to %d...\n", err, defaultTermWidth)
		termWidth = defaultTermWidth
	}

	/* Move the cursor up to overwrite previous progress output. */
	if lastPrintedLines > 0 {
		fmt.Printf("\x1b[%dA", lastPrintedLines) // ANSI escape: move cursor up N lines
	}

	linesCount := 0
	for _, title := range titles {
		processed := title.Images.GetAmtProcessed()
		total := title.Images.GetImgCount()

		leftText := getLeftText(title.Name, titleSpacing)
		rightText := getRightText(processed, total)

		line := formatLine(leftText, rightText, termWidth)
		fmt.Printf("\r%s\n", line)
		linesCount++

		for _, episode := range title.Episodes {
			processed := episode.Images.GetAmtProcessed()
			total := episode.Images.GetImgCount()

			leftText := getLeftText(episode.Name, episodeSpacing)
			rightText := getRightText(processed, total)

			line := formatLine(leftText, rightText, termWidth)
			fmt.Printf("\r%s\n", line)
			linesCount++
		}
	}

	lastPrintedLines = linesCount
}

/*
Formats a line to have `leftText`, `rightText` appear on the left and right sides
of a window, respectively. Spacing is determined by the width `width`.
*/
func formatLine(leftText string, rightText string, totalWidth int) string {
	spacing := max(totalWidth-len(leftText)-len(rightText), 1)

	return leftText + strings.Repeat(" ", spacing) + rightText
}

/*
Returns the string to be rendered at the left side of the progress bar.

Namely, the name of either a title or episode, given by `name`,
prefixed with `spacing` amount of whitespaces.
*/
func getLeftText(name string, spacing int) string {
	return strings.Repeat(" ", spacing) + name
}

/*
Returns the string to be rendered at the right side of the progress bar.

`processed` and `total` indicate the number of processed units and total units, respectively.
*/
func getRightText(processed uint32, total uint32) string {
	ratio := fmt.Sprintf("%*s", ratioWidth, fmt.Sprintf("(%d/%d)", processed, total))
	pbar := createProgressBar(processed, total)
	percentage := fmt.Sprintf("%*s", percentageWidth, fmt.Sprintf("%d%%", int(float64(processed)/float64(total)*100)))

	return strings.Join([]string{
		ratio,
		pbar,
		percentage,
	}, " ")
}

/*
Creates a simple progress bar, where `amtProcessed` is the number of units
processed so far and `total` is the total number of units.
*/
func createProgressBar(amtProcessed uint32, total uint32) string {
	completed := int(amtProcessed * progressWidth / total)
	remaining := int(progressWidth) - completed

	return "[" +
		strings.Repeat(saucer, completed) +
		strings.Repeat(saucerPadding, remaining) +
		"]"
}
