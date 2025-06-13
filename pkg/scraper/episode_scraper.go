package scraper

import (
	"fmt"
	"regexp"

	"github.com/gocolly/colly"
)

/* An episode of a title. */
type Episode struct {
	Name string
	Link string
}

/* Given a TV series title `title`, return its list of episodes. */
func (title Title) GetTVEpisodes() []Episode {
	var episodes []Episode

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
	)

	/*
		Extract the episode's name and link. (TV-only)
	*/
	c.OnHTML("h3 > a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		episode := Episode{
			Name: getEpisodeTitle(e.Text),
			Link: link,
		}
		episodes = append(episodes, episode)
	})

	/*
		If there is a next page,
		visit it and re-trigger episode name/link extraction. (TV-only)
	*/
	c.OnHTML("ul.pager > li > a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPageURL != "#" && containsNext(e.Text) {
			c.Visit(nextPageURL)
		}
	})

	/* Before making a request, print "Visiting: <URL>" */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector on the title. */
	c.Visit(title.Link)

	return episodes
}

/* Given an Anime title `title`, return its list of episodes. */
func (title Title) GetAnimeEpisodes() []Episode {
	var episodes []Episode

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
	)

	/*
		Extract the episode's name and link. (Anime-only)
	*/
	c.OnHTML("a[href] > h3", func(e *colly.HTMLElement) {
		href, _ := e.DOM.Parent().Attr("href")
		link := e.Request.AbsoluteURL(href)
		episode := Episode{
			Name: getEpisodeTitle(e.Text),
			Link: link,
		}
		episodes = append(episodes, episode)
	})

	/*
		If there is a next page,
		visit it and re-trigger episode name/link extraction. (Anime-only)
	*/
	c.OnHTML("a[title='Next Page']", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(nextPageURL)
	})

	/* Before making a request, print "Visiting: <URL>" */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector on the title. */
	c.Visit(title.Link)

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

/*
Returns true, if `s` contains "next" (case-insensitive).
Meant to be a heuristic to detect the next page for a TV Series.
*/
func containsNext(s string) bool {
	re := regexp.MustCompile(`(?i)next`)

	return re.MatchString(s)
}
