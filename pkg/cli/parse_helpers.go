package cli

import (
	"fmt"
	"os"
	"time"

	"sheeper.com/fancaps-scraper-go/pkg/fsutil"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

/*
Returns the output directory `outputDir` if its parent directories exist,
and exits with status code 1 otherwise.
*/
func validateOutputDir(outputDir string) string {
	if !fsutil.ParentDirsExist(outputDir) {
		fmt.Fprintf(os.Stderr,
			ui.ErrStyle.Render("couldn't find parent directories of `%s`")+"\n"+
				ui.ErrStyle.Render("make sure the parent directories exists.")+"\n",
			outputDir)
		os.Exit(1)
	}

	return outputDir
}

/*
Returns the amount of parallel downloads to make if it is strictly positive,
and exits with status code 1 otherwise.
*/
func validateParallelDownloads(parallelDownloads uint8) uint8 {
	if parallelDownloads == 0 {
		fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("parallel downloads must be strictly positive."))
		os.Exit(1)
	}

	return parallelDownloads
}

/*
Returns the delay time `delay` if it is non-negative,
and exits with status code 1 otherwise.
*/
func validateDelay(delay time.Duration) time.Duration {
	if delay < 0 {
		fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("delays must be non-negative."))
		os.Exit(1)
	}

	return delay
}
