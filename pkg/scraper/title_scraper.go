package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Given a URL `searchURL`, return all titles found by FanCaps. */
func GetTitles(searchURL string, flags cli.CLIFlags) []types.Title {
	var titles []types.Title

	/* Base options for the scraper. */
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains("fancaps.net"),
	}

	/* Enable asynchronous mode. */
	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(scraperOpts...)

	/* Extract the title's name and link. */
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		title := types.Title{
			Category: getCategory(link),
			Name:     e.Text,
			Link:     link,
		}
		titles = append(titles, title)
	})

	/* Suppress scraper output. */
	if flags.Quiet {
		/* Before making a request, print "Visiting: <URL>" */
		c.OnRequest(func(req *colly.Request) {
			fmt.Println("Visiting Search URL:", req.URL.String())
		})
	}

	/* Start the collector. */
	c.Visit(searchURL)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}

	/* Debug: Print found titles. */
	if flags.Debug {
		fmt.Println("FOUND TITLES:")
		for _, t := range titles {
			fmt.Println(t.Name, t.Link)
		}
		fmt.Println()
	}

	return titles
}

/* Return the category of a title based on its URL, `url`. */
func getCategory(url string) types.Category {
	switch {
	case strings.Contains(url, "/movies/"):
		return types.CategoryMovie
	case strings.Contains(url, "/tv/"):
		return types.CategoryTV
	case strings.Contains(url, "/anime/"):
		return types.CategoryAnime
	default:
		return types.CategoryUnknown
	}
}
