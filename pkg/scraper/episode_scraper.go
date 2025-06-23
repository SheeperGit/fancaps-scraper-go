package scraper

import (
	"fmt"
	"regexp"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Given a TV series title `title`, return its list of episodes. */
func GetTVEpisodes(title types.Title) []types.Episode {
	var episodes []types.Episode

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
	)

	/*
		Extract the episode's name and link. (TV-only)
	*/
	c.OnHTML("h3 > a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		episode := types.Episode{
			Name: getEpisodeTitle(e.Text),
			Link: link,
		}
		episodes = append(episodes, episode)
	})

	/*
		If there is a next page,
		visit it to re-trigger episode name/link extraction. (TV-only)
	*/
	c.OnHTML("ul.pager > li > a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPageURL != "#" && containsNext(e.Text) {
			c.Visit(nextPageURL)
		}
	})

	/* Before making a request, print "Visiting: <URL>" */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting TV Episode URL:", req.URL.String())
	})

	/* Start the collector on the title. */
	c.Visit(title.Link)

	return episodes
}

/* Given an Anime title `title`, return its list of episodes. */
func GetAnimeEpisodes(title types.Title) []types.Episode {
	var episodes []types.Episode

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
	)

	/* Extract the episode's name and link. (Anime-only) */
	c.OnHTML("a[href] > h3", func(e *colly.HTMLElement) {
		href, _ := e.DOM.Parent().Attr("href")
		link := e.Request.AbsoluteURL(href)
		episode := types.Episode{
			Name: getEpisodeTitle(e.Text),
			Link: link,
		}
		episodes = append(episodes, episode)
	})

	/*
		If there is a next page,
		visit it to re-trigger episode name/link extraction. (Anime-only)
	*/
	c.OnHTML("a[title='Next Page']", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(nextPageURL)
	})

	/* Before making a request, print "Visiting: <URL>" */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting Anime Episode URL:", req.URL.String())
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
