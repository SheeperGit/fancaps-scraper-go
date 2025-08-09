package scraper

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"golang.org/x/sync/errgroup"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
	"sheeper.com/fancaps-scraper-go/pkg/ui/prompt"
)

/*
Returns a list of search URLs to be scraped based on queries `queries` and categories `categories`.

If no queries have been specified, this function will prompt the user for queries and
validate them incrementally, and validates all queries in parallel otherwise.
*/
func GetSearchURLs(queries []string, categories []types.Category) []string {
	searchURLs := []string{}
	if len(queries) == 0 { // Prompt and validate search URLs incrementally.
		for len(queries) == 0 || prompt.YesNoPrompt("Enter another query? [y/N]: ", "") {
			query := prompt.TextPrompt("Enter Search Query: ", prompt.QueryHelpPrompt)
			if strings.TrimSpace(query) == "" {
				fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("search query cannot be empty.\n\n"))
				continue
			}

			url := BuildQueryURL(query, categories)
			if !titleExists(url) {
				fmt.Fprintf(os.Stderr, ui.ErrStyle.Render("no titles found for query `%s`.")+"\n\n", query)
				continue
			}
			fmt.Printf(ui.SuccessStyle.Render("Found titles for query: `%s`")+"\n", query)
			searchURLs = append(searchURLs, url)
			queries = append(queries, query)
		}
	} else { // Validate search URLs all at once.
		for _, query := range queries {
			if strings.TrimSpace(query) == "" { // fancaps.net considers empty queries as valid and returns a massive list otherwise.
				fmt.Fprintln(os.Stderr, "search query cannot be empty.")
				os.Exit(1)
			}
			url := BuildQueryURL(query, categories)
			searchURLs = append(searchURLs, url)
		}

		var eg errgroup.Group
		for i, url := range searchURLs {
			i, url := i, url // https://golang.org/doc/faq#closures_and_goroutines
			eg.Go(func() error {
				if !titleExists(url) {
					return fmt.Errorf("no titles found for query `%s`", queries[i])
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	return searchURLs
}

/*
Returns a URL which will be used to scrape titles using query `query`,
searching only categories in `categories`.
*/
func BuildQueryURL(query string, categories []types.Category) string {
	params := url.Values{}
	params.Add("q", query)

	for _, cat := range categories {
		switch cat {
		case types.CategoryMovie:
			params.Add("MoviesCB", "Movies")
		case types.CategoryTV:
			params.Add("TVCB", "TV")
		case types.CategoryAnime:
			params.Add("animeCB", "Anime")
		}
	}
	params.Add("submit", "Submit Query")

	return "https://fancaps.net/search.php" + "?" + params.Encode()
}

/* Returns true, if a title exists in the URL `searchURL`, and returns false otherwise. */
func titleExists(searchURL string) bool {
	titleExists := false
	flags := cli.Flags()

	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains("fancaps.net"),
	}

	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	c := colly.NewCollector(scraperOpts...)

	/* Search the results of each category. */
	c.OnHTML("div.single_post_content > table", func(e *colly.HTMLElement) {
		/* Title found. */
		e.ForEachWithBreak("h4 > a", func(_ int, _ *colly.HTMLElement) bool {
			titleExists = true
			return false // Stop searching for more titles.
		})
	})

	c.Visit(searchURL)

	if flags.Async {
		c.Wait()
	}

	return titleExists
}
