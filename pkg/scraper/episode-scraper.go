package scraper

import (
	"fmt"
	"regexp"

	"github.com/gocolly/colly"
)

type Episode struct {
	Name string
	Link string
}

func GetEpisodes(titles []Title) []Episode {
	var episodes []Episode

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),

		/* Disabled, as this doesn't guarantee in-order view */
		// colly.Async(true),
	)

	/*
		On every `a`` element which has an href attribute with an h3 child element,
		extract the episode name and a link to view the episode's images.
	*/
	c.OnHTML("a[href] > h3", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.DOM.Parent().AttrOr("href", ""))
		episode := Episode{
			Name: getEpisodeTitle(e.Text),
			Link: link,
		}
		episodes = append(episodes, episode)
	})

	/* If there is a next page, visit it and re-trigger episode name/link extraction. */
	c.OnHTML("a[title='Next Page']", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(nextPageURL)
	})

	/* Before making a request, print "Visiting ..." */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector on each of the titles. */
	for _, t := range titles {
		c.Visit(t.Link)
	}

	return episodes
}

/* Returns the episode's title. */
func getEpisodeTitle(baseTitle string) string {
	re := regexp.MustCompile("Images From (.+)")
	episodeTitle := re.FindStringSubmatch(baseTitle)
	if len(episodeTitle) >= 2 {
		return episodeTitle[1]
	}

	return "EPISODE TITLE NOT FOUND"
}
