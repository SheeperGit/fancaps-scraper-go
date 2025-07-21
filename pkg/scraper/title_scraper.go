package scraper

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/*
Returns a non-empty list of titles found through the URL `searchURL`.

This function re-prompts the user for a search query if the query flag was passed
and exits with code 1 otherwise.
*/
func GetTitles(searchURL string, flags cli.CLIFlags) []*types.Title {
	titles := scrapeTitles(searchURL, flags)
	for titles == nil {
		/* Exit if the query was passed as a CLA. */
		if flags.QueryCLAPassed {
			fmt.Fprintf(os.Stderr, "No titles found for query '%s'.\n", flags.Query)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Couldn't find any titles matching the query '%s'.\n", flags.Query)
		fmt.Fprintf(os.Stderr, "Try again with a different query.\n\n")

		/* Redo. */
		flags = cli.ParseCLI()
		searchURL = cli.BuildQueryURL(flags.Query, flags.Categories)
		titles = scrapeTitles(searchURL, flags)
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
		fmt.Println("\n\nFOUND TITLES:")
		for _, t := range titles {
			fmt.Println(t.Name, t.Link)
		}
		fmt.Printf("\n\n")
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
