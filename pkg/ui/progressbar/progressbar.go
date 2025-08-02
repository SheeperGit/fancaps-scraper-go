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
		ratioWidth = 2*len(strconv.Itoa(int(types.ImgTotal()))) + 3
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
			processed := title.Images.Processed()
			skipped := title.Images.Skipped()
			total := title.Images.Total()

			leftText := getLeftText(title.Name, titleSpacing)
			rightText := getRightText(processed, skipped, total, title.Start)

			line := formatLine(leftText, rightText, processed, total, termWidth, title.Images, nil)
			fmt.Printf("\r%s\n", line)
		} else {
			fmt.Println()
		}
		linesCount++

		for _, episode := range title.Episodes {
			/* Render episode progress, if not done. */
			if !episode.Images.Done {
				processed := episode.Images.Processed()
				skipped := episode.Images.Skipped()
				total := episode.Images.Total()

				baseEpisodeName := getBaseEpisodeName(episode)
				leftText := getLeftText(baseEpisodeName, episodeSpacing)
				rightText := getRightText(processed, skipped, total, episode.Start)

				line := formatLine(leftText, rightText, processed, total, termWidth, title.Images, episode.Images)
				fmt.Printf("\r%s\n", line)
			} else {
				fmt.Println()
			}
			linesCount++
		}
	}
	totalProcessed := types.ProcessedTotal()
	totalSkipped := types.SkippedTotal()
	totalImgs := types.ImgTotal()

	leftText := getLeftText("Total: ", totalSpacing)
	rightText := getRightText(totalProcessed, totalSkipped, totalImgs, downloadStart)

	line := formatLine(leftText, rightText, totalProcessed, totalImgs, termWidth, nil, nil)
	fmt.Printf("\r\n%s\n", line)
	linesCount = linesCount + 2

	lastPrintedLines = linesCount
}

/*
Updates and shows the progress of titles `titles`.

Always increments the progress of the title images `titleImages` (mandatory),
and may also increment the progress of the episode images `episodeImages` if non-nil (optional).
*/
func UpdateProgressDisplay(titles []*types.Title, titleImages *types.Images, episodeImages *types.Images) {
	if episodeImages != nil {
		episodeImages.IncrementProcessed()
	}
	titleImages.IncrementProcessed()

	ShowProgress(titles)
}

/*
Formats a line to have `leftText`, `rightText` appear on the left and right sides
of a window, respectively. Spacing is determined by the width `width`. Style is
determined by the amount of processed and total units, `processed`, `total`, respectively.
*/
func formatLine(leftText, rightText string, processed, total uint32, totalWidth int, titleImages, episodeImages *types.Images) string {
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
			/* If an episode was processed, check if its title was processed to mark it too. */
			episodeImages.Done = true
			if titleImages.Processed() == titleImages.Total() {
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
func getRightText(processed, skipped, total uint32, start time.Time) string {
	eta := getETAString(processed, skipped, total, start)
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

/*
Returns an ETA based on the start time `start`, and the number of processed
and total units, `processed`, `total`, respectively.

Returns an empty string if `processed` is 0.
*/
func getETAString(processed, skipped, total uint32, start time.Time) string {
	if total == processed {
		return ""
	}

	downloaded := processed - skipped

	/*
		If no previous download data available for a title/episode,
		estimate using global download data.
	*/
	if downloaded == 0 {
		globalDownloaded := types.ProcessedTotal() - types.SkippedTotal()
		if globalDownloaded == 0 {
			return "0s/--" // No download data available. No estimate!
		}
		globalElapsed := time.Since(downloadStart)
		globalRate := float64(globalElapsed) / float64(globalDownloaded)
		globalRemaining := time.Duration(globalRate * float64(total-processed)).Round(time.Second)
		remainingWidth := len(globalRemaining.String())

		return fmt.Sprintf("%*s", remainingWidth+5, fmt.Sprintf("(0s/%s)", globalRemaining))
	}

	/* Otherwise, use local title/episode download estimate. */
	elapsed := time.Since(start)
	rate := float64(elapsed) / float64(downloaded)
	remaining := time.Duration(rate * float64(total-processed))

	elapsed = elapsed.Round(time.Second)
	elapsedWidth := len(elapsed.String())
	remaining = remaining.Round(time.Second)
	remainingWidth := len(remaining.String())

	return fmt.Sprintf("%*s", elapsedWidth+remainingWidth+3, fmt.Sprintf("(%s/%s)", elapsed, remaining))
}
