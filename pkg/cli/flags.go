package cli

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"sheeper.com/fancaps-scraper-go/pkg/menu"
)

type CLIFlags struct {
	Query  string
	Movies bool
	TV     bool
	Anime  bool
}

/*
Parse CLI flags.
Always returns non-empty Query.
*/
func ParseCLI() CLIFlags {
	query := flag.String("q", "", "Search query term")
	movies := flag.Bool("movies", false, "Include Movies in search query")
	tv := flag.Bool("tv", false, "Include TV series in search query")
	anime := flag.Bool("anime", false, "Include Anime in search query")

	flag.Parse()

	/* If `-q` not specified, prompt user for search query. */
	for *query == "" {
		*query = getSearchQuery()
		if *query == "" {
			fmt.Fprintf(os.Stderr, "CLI Error: Search query cannot be empty.\n")
			flag.Usage()
		}
	}

	/* If no categories flags specified, open Category Menu. */
	if !*movies && !*tv && !*anime {
		selectedMenuCategories, confirmed := menu.GetCategoriesMenu()
		if !confirmed {
			fmt.Fprintf(os.Stderr, "Category Menu: Operation aborted.\n")
			os.Exit(1)
		}

		/* Set active categories according to Category Menu. */
		for cat := range selectedMenuCategories {
			switch cat {
			case menu.MOVIE_TEXT:
				*movies = true
			case menu.TV_TEXT:
				*tv = true
			case menu.ANIME_TEXT:
				*anime = true
			}
		}
	}

	return CLIFlags{
		Query:  *query,
		Movies: *movies,
		TV:     *tv,
		Anime:  *anime,
	}
}

/*
Returns initial URL to scrape based on search query, `CLIFlags.Query`.
*/
func (f CLIFlags) BuildQueryURL() string {
	params := url.Values{}
	params.Add("q", f.Query)
	if f.Movies {
		params.Add("MoviesCB", "Movies")
	}
	if f.TV {
		params.Add("TVCB", "TV")
	}
	if f.Anime {
		params.Add("animeCB", "Anime")
	}
	params.Add("submit", "Submit Query")

	const baseURL = "https://fancaps.net/search.php"
	return baseURL + "?" + params.Encode()
}

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
