package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Given a URL `searchURL`, return all titles found by FanCaps. */
func GetTitles(searchURL string) []types.Title {
	var titles []types.Title

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
	)

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

	/* Before making a request, print "Visiting: <URL>" */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting Search URL:", req.URL.String())
	})

	/* Start the collector. */
	c.Visit(searchURL)

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
