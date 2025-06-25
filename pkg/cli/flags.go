package cli

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"slices"

	"sheeper.com/fancaps-scraper-go/pkg/menu"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Available CLI Flags. */
type CLIFlags struct {
	Query      string           // Search query to scrape from.
	Categories []types.Category // Selected categories to search using `Query`.
	Async      bool             // If true, enable asynchronous network requests.
	Debug      bool             // If true, print final selections and scraped results after completion.
	Quiet      bool             // If true, suppress scraper output.
}

/*
Parse CLI flags.
Always returns non-empty Query.
*/
func ParseCLI() CLIFlags {
	/* Query Flags. */
	query := flag.String("q", "", "Search query term")

	/* Category Flags. */
	movies := flag.Bool("movies", false, "Include Movies in search query")
	tv := flag.Bool("tv", false, "Include TV series in search query")
	anime := flag.Bool("anime", false, "Include Anime in search query")

	/* Optimization Flags. */
	async := flag.Bool("async", true, "Enable asynchronous requests (recommended: significantly improves speed on most systems)")

	/* Miscellaneous Flags. */
	debug := flag.Bool("debug", false, "Enable debug mode (print final selections and scraped results after completion)")
	quiet := flag.Bool("quiet", false, "Suppress scraper output")

	flag.Parse()

	/* If `-q` not specified, prompt user for search query. */
	for *query == "" {
		*query = getSearchQuery()
		if *query == "" {
			fmt.Fprintf(os.Stderr, "CLI Error: Search query cannot be empty.\n")
			flag.Usage()
		}
	}

	var categories []types.Category
	if *anime {
		categories = append(categories, types.CategoryAnime)
	}
	if *tv {
		categories = append(categories, types.CategoryTV)
	}
	if *movies {
		categories = append(categories, types.CategoryMovie)
	}

	/* If no categories flags specified, open Category Menu. */
	if len(categories) == 0 {
		selectedMenuCategories := menu.LaunchCategoriesMenu()

		/* Set active categories according to Category Menu. */
		for cat := range selectedMenuCategories {
			categories = append(categories, cat)
		}

		/* Sort according to Category enum order. */
		slices.Sort(categories)
	}

	return CLIFlags{
		Query:      *query,
		Categories: categories,
		Async:      *async,
		Debug:      *debug,
		Quiet:      *quiet,
	}
}

/* Returns initial URL to scrape based on search query, `CLIFlags.Query`. */
func (flags CLIFlags) BuildQueryURL() string {
	params := url.Values{}
	params.Add("q", flags.Query)

	for _, cat := range flags.Categories {
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

	const baseURL = "https://fancaps.net/search.php"
	return baseURL + "?" + params.Encode()
}

/* Prompt user for a search query to scrape titles from. */
func getSearchQuery() string {
	fmt.Print("Enter Search Query: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}
