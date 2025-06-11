package scraper

import (
	"fmt"

	"github.com/gocolly/colly"
)

type Title struct {
	Name string
	Link string
}

func GetTitles(searchURL string) []Title {
	var titles []Title

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
		colly.Async(true),
	)

	/*
		On every h4 element which has an anchor child element,
		print the link text and the link itself.
	*/
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		title := Title{
			Name: e.Text,
			Link: link,
		}
		titles = append(titles, title)
	})

	/* Before making a request, print "Visiting ..." */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector. */
	c.Visit(searchURL)

	/* Wait for the collector to finish. (Required for Async) */
	c.Wait()

	return titles
}
