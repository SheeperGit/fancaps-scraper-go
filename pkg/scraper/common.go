package scraper

import (
	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
)

const AllowedDomains = "fancaps.net" // Domains the scraper is allowed to visit.

func GetScraperOpts(flags cli.CLIFlags) []func(*colly.Collector) {
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains(AllowedDomains),
	}

	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	return scraperOpts
}
