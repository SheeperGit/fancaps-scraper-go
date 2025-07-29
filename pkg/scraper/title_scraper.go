package scraper

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

/*
Returns a non-empty, unique, sorted list of titles found through the URLs `searchURLs`,
which are assumed to be validated (have at least one title associated with each URL).
*/
func GetTitles(searchURLs []string, flags cli.CLIFlags) []*types.Title {
	var (
		titles   []*types.Title              // Scraped titles.
		titlesMu sync.Mutex                  // Prevents overlapping "appends" to `titles`.
		wg       sync.WaitGroup              // Synchronizes title scrapers.
		seen     = make(map[string]struct{}) // Duplicate titles protection.
	)

	for _, searchURL := range searchURLs {
		wg.Add(1)
		go func(searchURL string) {
			defer wg.Done()

			ts := scrapeTitles(searchURL, flags)

			titlesMu.Lock()
			for _, t := range ts {
				if _, exists := seen[t.Link]; !exists {
					seen[t.Link] = struct{}{}
					titles = append(titles, t)
				}
			}
			titlesMu.Unlock()
		}(searchURL)
	}
	wg.Wait()

	/* Sort found titles by category, then alphabetically. (Case-insensitive) */
	sort.Slice(titles, func(i, j int) bool {
		catI := titles[i].Category
		catJ := titles[j].Category

		if catI != catJ {
			return catI < catJ
		}

		return strings.ToLower(titles[i].Name) < strings.ToLower(titles[j].Name)
	})

	/* Debug: Print found titles. */
	if flags.Debug {
		_, maxTitleWidth := ui.GetLongestTitle(titles)

		fmt.Println("\n\nFOUND TITLES:")
		for _, t := range titles {
			fmt.Printf("%-*s %s\n", maxTitleWidth, t.Name, t.Link)
		}
		fmt.Printf("\n\n")
	}

	return titles
}

/* Given a URL `searchURL`, return all titles found by FanCaps. */
func scrapeTitles(searchURL string, flags cli.CLIFlags) []*types.Title {
	var titles []*types.Title

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
		category := getCategory(link)
		title := &types.Title{
			Category: category,
			Name:     e.Text,
			Link:     link,
			Images:   &types.Images{},
		}
		titles = append(titles, title)
	})

	/* Suppress scraper output. */
	if flags.Debug {
		/* Before making a request, print "Visiting: <URL>" */
		c.OnRequest(func(req *colly.Request) {
			fmt.Printf("SEARCH QUERY URL: %s\n", req.URL.String())
		})
	}

	/* Start the collector. */
	c.Visit(searchURL)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
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
		fmt.Fprintf(os.Stderr, "getCategory: couldn't extract category from url %s", url)
		os.Exit(1)
		return -1
	}
}
