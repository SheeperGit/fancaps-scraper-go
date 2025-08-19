package scraper

import (
	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
)

const allowedDomains = "fancaps.net" // Domains the scraper is allowed to visit.

/* Returns the scraper options from flags `flags`. */
func GetScraperOpts(flags cli.CLIFlags) []func(*colly.Collector) {
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains(allowedDomains),
	}

	if !flags.NoAsync {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	return scraperOpts
}
