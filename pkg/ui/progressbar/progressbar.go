package progressbar

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

const (
	saucer        = "#" // Character indicating complete progress.
	saucerPadding = "-" // Character indicating incomplete progress.
)

const (
	defaultTermWidth = 80 // Default terminal width. (Fallback)
	progressbarWidth = 52 // Amount of glyphs within the progress bar.
	percentageWidth  = 4  // Width taken by the percentage of completed/total images downloaded in the progress bar.
)

const (
	titleSpacing   = 1 // Spacing applied to title names on a progress line.
	episodeSpacing = 3 // Spacing applied to episode names on a progress line.
	totalSpacing   = 0 // Spacing applied to the total progress line.
)

var (
	progressMu       sync.Mutex // Limits progress bar access to one thread.
	lastPrintedLines int        // Number of lines last printed by the progress display.
)

var (
	setOnce       sync.Once // Initializes certain progress variables.
	downloadStart time.Time // Timestamp marking the start of the image download process for all titles.
	ratioWidth    int       // Width taken by the ratio of completed/total images downloaded in the progress bar.
)

/* Displays progress bar(s) based on the state of the titles `titles`. */
func ShowProgress(titles []*types.Title) {
	setOnce.Do(func() {
		downloadStart = time.Now()
		ratioWidth = 2*len(strconv.Itoa(int(types.GlobalTotalImages()))) + 3
	})

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
		/* Render title progress, if not done. */
		if !title.Images.Done {
			downloaded := title.Images.Downloaded()
			skipped := title.Images.Skipped()
			total := title.Images.Total()

			leftText := getLeftText(title.Name, titleSpacing)
			rightText := getRightText(downloaded, skipped, total, title.Start)

			line := formatLine(leftText, rightText, downloaded, skipped, total, termWidth, title.Images, nil)
			fmt.Printf("\r%s\n", line)
		} else {
			fmt.Println()
		}
		linesCount++

		for _, episode := range title.Episodes {
			/* Render episode progress, if not done. */
			if !episode.Images.Done {
				downloaded := episode.Images.Downloaded()
				skipped := episode.Images.Skipped()
				total := episode.Images.Total()

				baseEpisodeName := getBaseEpisodeName(episode)
				leftText := getLeftText(baseEpisodeName, episodeSpacing)
				rightText := getRightText(downloaded, skipped, total, episode.Start)

				line := formatLine(leftText, rightText, downloaded, skipped, total, termWidth, title.Images, episode.Images)
				fmt.Printf("\r%s\n", line)
			} else {
				fmt.Println()
			}
			linesCount++
		}
	}
	totalDownloaded := types.GlobalDownloadedImages()
	totalSkipped := types.GlobalSkippedImages()
	totalImgs := types.GlobalTotalImages()

	leftText := getLeftText("Total: ", totalSpacing)
	rightText := getRightText(totalDownloaded, totalSkipped, totalImgs, downloadStart)

	line := formatLine(leftText, rightText, totalDownloaded, totalSkipped, totalImgs, termWidth, nil, nil)
	fmt.Printf("\r\n%s\n", line)
	linesCount = linesCount + 2

	lastPrintedLines = linesCount
}

/*
Increments the progress of an image container using incrementer function `incFunc`,
and shows the progress of titles `titles`.
*/
func UpdateProgressDisplay(titles []*types.Title, incFunc func()) {
	incFunc()
	ShowProgress(titles)
}

/*
Formats a line to have `leftText`, `rightText` appear on the left and right sides
of a window, respectively. Spacing is determined by the width `width`. Style is
determined by the amount of downloaded, skipped, and total units, `downloaded`,
`skipped`, `total`, respectively.
*/
func formatLine(leftText, rightText string, download, skipped, total uint32, totalWidth int, titleImages, episodeImages *types.Images) string {
	processed := download + skipped
	spacing := max(totalWidth-len(leftText)-len(rightText), 1)

	line := ""
	switch {
	case processed == 0:
		line = leftText + strings.Repeat(" ", spacing) + rightText
	case processed < total:
		line = ui.HighlightStyle.Render(leftText + strings.Repeat(" ", spacing) + rightText)
	case processed == total:
		line = ui.SuccessStyle.Render(leftText + strings.Repeat(" ", spacing) + rightText)

		switch {
		case titleImages == nil:
			// Do nothing. (We're on the total progress line, so we're done rendering!)
		case episodeImages == nil:
			/* If all title images processed, mark it to skip future renders. */
			titleImages.Done = true
		default:
			/* If an episode was fully processed, check if its title was processed to mark it too. */
			episodeImages.Done = true
			if titleImages.Downloaded()+titleImages.Skipped() == titleImages.Total() {
				titleImages.Done = true
			}
		}
	}

	return line
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

`processed`, `total`, `start` indicate the number of processed units, total units, and
start time of the title or episode, respectively.
*/
func getRightText(downloaded, skipped, total uint32, start time.Time) string {
	processed := downloaded + skipped

	eta := getETAString(downloaded, skipped, total, start)
	ratio := fmt.Sprintf("%*s", ratioWidth, fmt.Sprintf("(%d/%d)", processed, total))
	pbar := createProgressBar(processed, total)
	percentage := fmt.Sprintf("%*s", percentageWidth, fmt.Sprintf("%d%%", int(float64(processed)/float64(total)*100)))

	return strings.Join([]string{
		eta,
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
	completed := int(amtProcessed * progressbarWidth / total)
	remaining := int(progressbarWidth) - completed

	return "[" +
		strings.Repeat(saucer, completed) +
		strings.Repeat(saucerPadding, remaining) +
		"]"
}

/*
Returns an ETA based on the start time `start`, and the number of downloaded, skipped,
and total units, `downloaded`, `skipped`, `total`, respectively.
*/
func getETAString(downloaded, skipped, total uint32, start time.Time) string {
	/* If no previous download data available, estimate using global download data. */
	if downloaded == 0 {
		globalDownloaded := types.GlobalDownloadedImages()
		if globalDownloaded == 0 {
			return "0s/--" // No download data available. No estimate!
		}

		globalElapsed := time.Since(downloadStart)
		globalRate := float64(globalElapsed) / float64(globalDownloaded)
		globalRemaining := time.Duration(globalRate * float64(total-downloaded-skipped)).Round(time.Second)
		remainingWidth := len(globalRemaining.String())

		return fmt.Sprintf("%*s", remainingWidth+5, fmt.Sprintf("(0s/%s)", globalRemaining))
	}

	/* Otherwise, use local download data to estimate. */
	elapsed := time.Since(start)
	rate := float64(elapsed) / float64(downloaded)
	remaining := time.Duration(rate * float64(total-downloaded-skipped)).Round(time.Second)
	elapsed = elapsed.Round(time.Second)

	elapsedWidth := len(elapsed.String())
	remainingWidth := len(remaining.String())

	return fmt.Sprintf("%*s", elapsedWidth+remainingWidth+3, fmt.Sprintf("(%s/%s)", elapsed, remaining))
}

/*
Returns the base episode name of episode `episode`.
If the base name is unable to be extracted for whatever reason,
the original name is returned.

For example, "Episode 2 of Neon Genesis Evangelion" -> "Episode 2"
*/
func getBaseEpisodeName(episode *types.Episode) string {
	re := regexp.MustCompile(`^(.*?)\s+of\b`)

	matches := re.FindStringSubmatch(episode.Name)
	if len(matches) > 1 {
		return matches[1]
	}

	return episode.Name
}
