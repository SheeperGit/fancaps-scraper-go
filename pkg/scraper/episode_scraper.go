package scraper

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Get episodes from titles `titles`. */
func GetEpisodes(titles []*types.Title, flags cli.CLIFlags) []*types.Title {
	var wg sync.WaitGroup

	if flags.Debug {
		fmt.Println("\nEPISODE LINKS VISITED:")
	}

	/* Get the episodes for each title. */
	for _, title := range titles {
		/*
			From title category, run corresponding episode scraper.
			Note: Movies do not have episodes and thus do not require episode scraping.
		*/
		scrapeEpisodes := func(t *types.Title) {
			switch t.Category {
			case types.CategoryAnime:
				t.Episodes = GetAnimeEpisodes(t, flags)
			case types.CategoryTV:
				t.Episodes = GetTVEpisodes(t, flags)
			case types.CategoryMovie:
				// Do nothing
			default:
				fmt.Fprintf(os.Stderr, "Unknown Category: %s (%s) -> [%s]\n", title.Name, title.Link, title.Category)
			}
		}

		if flags.Async {
			wg.Add(1)
			go func(title *types.Title) {
				defer wg.Done()
				scrapeEpisodes(title)
			}(title)
		} else {
			scrapeEpisodes(title)
		}
	}

	if flags.Async {
		wg.Wait()
	}

	/* Debug: Print found titles and episodes. */
	if flags.Debug {
		fmt.Println("\n\nFOUND TITLES AND EPISODES:")
		for _, title := range titles {
			fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
			for _, episode := range title.Episodes {
				fmt.Printf("\t%s -> %s\n", episode.Name, episode.Link)
			}
		}
	}

	return titles
}

/* Given a TV series title `title`, return its list of episodes. */
func GetTVEpisodes(title *types.Title, flags cli.CLIFlags) []*types.Episode {
	var episodes []*types.Episode

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

	/* Extract the episode's name and link. (TV-only) */
	c.OnHTML("h3 > a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		episode := &types.Episode{
			Name:   getEpisodeTitle(e.Text),
			Link:   link,
			Images: &types.Images{},
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

	/* Suppress scraper output. */
	if flags.Debug {
		c.OnRequest(func(req *colly.Request) {
			fmt.Println("Visiting TV Episode URL:", req.URL.String())
		})
	}

	/* Start the collector on the title. */
	c.Visit(title.Link)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}

	return episodes
}

/* Given an Anime title `title`, return its list of episodes. */
func GetAnimeEpisodes(title *types.Title, flags cli.CLIFlags) []*types.Episode {
	var episodes []*types.Episode

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

	/* Extract the episode's name and link. (Anime-only) */
	c.OnHTML("a[href] > h3", func(e *colly.HTMLElement) {
		href, _ := e.DOM.Parent().Attr("href")
		link := e.Request.AbsoluteURL(href)
		episode := &types.Episode{
			Name:   getEpisodeTitle(e.Text) + " of " + title.Name, // Append title name (required for `getEpisodeByNumber()`)
			Link:   link,
			Images: &types.Images{},
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

	/* Suppress scraper output. */
	if flags.Debug {
		c.OnRequest(func(req *colly.Request) {
			fmt.Println("Visiting Anime Episode URL:", req.URL.String())
		})
	}

	/* Start the collector on the title. */
	c.Visit(title.Link)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}

	return episodes
}

/* Returns the episode's title. */
func getEpisodeTitle(baseTitle string) string {
	re := regexp.MustCompile(`Images From (.+?)\s*$`)
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
