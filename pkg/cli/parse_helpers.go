package cli

import (
	"fmt"
	"os"
	"time"

	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

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
