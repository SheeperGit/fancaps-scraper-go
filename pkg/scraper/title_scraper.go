package scraper

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/menu"
)

type Title struct {
	Category menu.Category
	Episodes []Episode
	Name     string
	Link     string
}

func GetTitles(searchURL string) []Title {
	var titles []Title

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
	)

	/*
		On every h4 element which has an anchor child element,
		extract the title name and the link to view the title's episode(s), if any.
	*/
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		title := Title{
			Category: getCategory(link),
			Name:     e.Text,
			Link:     link,
		}
		titles = append(titles, title)
	})

	/* Before making a request, print "Visiting: <URL>" */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector. */
	c.Visit(searchURL)

	return titles
}

/* Return the category of a title based on its URL. */
func getCategory(url string) menu.Category {
	switch {
	case strings.Contains(url, "/movies/"):
		return menu.CategoryMovie
	case strings.Contains(url, "/tv/"):
		return menu.CategoryTV
	case strings.Contains(url, "/anime/"):
		return menu.CategoryAnime
	default:
		return menu.CategoryUnknown
	}
}
