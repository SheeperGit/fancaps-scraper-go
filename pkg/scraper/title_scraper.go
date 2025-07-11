package scraper

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Given a URL `searchURL`, return all titles found by FanCaps. */
func GetTitles(searchURL string, catStats *types.CatStats, flags cli.CLIFlags) []types.Title {
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
		category := getCategory(link)
		title := types.Title{
			Category: category,
			Name:     e.Text,
			Link:     link,
		}
		titles = append(titles, title)
		catStats.Increment(category)
	})

	/* Suppress scraper output. */
	if flags.Debug {
		/* Before making a request, print "Visiting: <URL>" */
		c.OnRequest(func(req *colly.Request) {
			fmt.Printf("SEARCH QUERY URL: %s\n\n", req.URL.String())
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

	/* Debug: Print category statistics and found titles. */
	if flags.Debug {

		snapshot := catStats.Snapshot()

		var categories []types.Category
		for cat := range snapshot {
			categories = append(categories, cat)
		}
		slices.Sort(categories)

		fmt.Println("CATEGORY STATISTICS:")
		for _, cat := range categories {
			fmt.Printf("\t%s Found: %d\n", cat.String(), snapshot[cat])
		}
		fmt.Println()

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
		fmt.Fprintf(os.Stderr, "getCategory: couldn't extract category from url %s", url)
		os.Exit(1)
		return -1
	}
}
