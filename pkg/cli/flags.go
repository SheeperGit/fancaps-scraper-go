package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gocolly/colly"
	"github.com/spf13/cobra"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
	"sheeper.com/fancaps-scraper-go/pkg/ui/prompt"
)

/* Available CLI Flags. */
type CLIFlags struct {
	Query      string           // Search query to scrape from.
	Categories []types.Category // Selected categories to search using `Query`.
	OutputDir  string           // The directory to output images.
	Async      bool             // If true, enable asynchronous network requests.
	Debug      bool             // If true, print useful debugging messages.
}

/* Example usage of fancaps-scraper-go. */
const exampleUsage = `
  # Show this message and exit.
  fancaps-scraper --help

  # Search for "Naruto" with anime and tv series titles only.
  fancaps-scraper --query Naruto --categories anime,tv

  # Search for "The Office" (with short flags) in all categories. (Notice also the single quotes to signify "The Office" as one argument.)
  fancaps-scraper -q 'The Office' -c all

  # Search for "Inception" movie titles only, with debug enabled.
  fancaps-scraper -q Inception --categories movies --debug

  # Search for "Friends" tv series titles only, with asynchronous network requests explicitly disabled.
  fancaps-scraper -q Friends --categories tv --async=false`

var defaultOutputDir = filepath.Join(".", "output") // Default output directory.

/*
Parse CLI flags.
Always returns non-empty Query.
*/
func ParseCLI() CLIFlags {
	var (
		flags      CLIFlags
		query      string
		categories string
		outputDir  string
		async      bool
		debug      bool
	)

	rootCmd := &cobra.Command{
		Use:     "fancaps-scraper",
		Short:   "Scrape images from fancaps.net using a CLI interface",
		Example: exampleUsage,
		Run: func(cmd *cobra.Command, args []string) {
			/* Check that the parent directories exist. */
			if !ParentDirsExist(outputDir) {
				fmt.Fprintf(os.Stderr, "ParseCLI error: Couldn't find parent directories of '%s'\n", outputDir)
				fmt.Fprintf(os.Stderr, "Make sure the parent directories exists.\n")
				os.Exit(1)
			}
			flags.OutputDir = outputDir

			/* Category Parsing. */
			if categories != "" {
				sanitizedInput := strings.ToLower(categories)
				parts := strings.Split(sanitizedInput, ",")

				categoryMap := map[string]types.Category{
					"anime":  types.CategoryAnime,
					"tv":     types.CategoryTV,
					"movies": types.CategoryMovie,
				}

				seen := map[types.Category]bool{}

				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part == "all" {
						for _, cat := range categoryMap {
							if !seen[cat] {
								flags.Categories = append(flags.Categories, cat)
								seen[cat] = true
							}
						}
						break
					}

					if cat, ok := categoryMap[part]; ok && !seen[cat] {
						flags.Categories = append(flags.Categories, cat)
						seen[cat] = true
					} else if !ok {
						fmt.Fprintf(os.Stderr, "CLI Error: Unknown category '%s'. Valid options are: anime, tv, movies, all\n", part)
						os.Exit(1)
					}
				}
			}

			/* If no categories flags specified, open Category Menu. */
			if len(flags.Categories) == 0 {
				selectedMenuCategories := menu.LaunchCategoriesMenu()
				for cat := range selectedMenuCategories {
					flags.Categories = append(flags.Categories, cat)
				}
			}

			/* Sort according to Category enum order. */
			slices.Sort(flags.Categories)

			/* If `-q` was specified, exit if no titles exist for the query. */
			searchURL := BuildQueryURL(query, flags.Categories)
			if cmd.Flags().Changed("query") && !titleExists(searchURL, flags) {
				fmt.Fprintf(os.Stderr, "No titles found for query '%s'.\n", query)
				os.Exit(1)
			}

			/* If `-q` not specified, prompt user for search query. */
			for query == "" {
				query = prompt.PromptUser("Enter Search Query: ", prompt.SearchHelpPrompt)
				if query == "" {
					fmt.Fprintln(os.Stderr, "CLI Error: Search query cannot be empty.")
					cmd.Usage()
					os.Exit(1)
				}
			}

			flags.Query = query
			flags.Async = async
			flags.Debug = debug
		},
	}

	/* Flag Definitions. */
	rootCmd.Flags().StringVarP(&query, "query", "q", "", "Search query term")
	rootCmd.Flags().StringVarP(&categories, "categories", "c", "", "Categories to search. Format: [anime,tv,movies|all] (comma-separated)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", defaultOutputDir, "Output directory for images. (Parent directories must exist)")
	rootCmd.Flags().BoolVar(&async, "async", true, "Enable asynchronous requests")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")

	/* "Override" default help. */
	rootCmd.Flags().BoolP("help", "h", false, "Display this help and exit")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Root().Usage()
		os.Exit(0)
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return flags
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

	const baseURL = "https://fancaps.net/search.php"
	return baseURL + "?" + params.Encode()
}

/*
Returns true, if the parent directories of `dirPath` exist
and returns false otherwise.
*/
func ParentDirsExist(dirPath string) bool {
	parentDirs := filepath.Dir(dirPath)

	info, err := os.Stat(parentDirs)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Fprintf(os.Stderr, "ParentDirsExist unexpected error: %v", err)
		return false
	}

	return info.IsDir()
}

/* Returns true, if a title exists in the URL `searchURL`, and returns false otherwise. */
func titleExists(searchURL string, flags CLIFlags) bool {
	titleExists := false

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

	/* At least one title was found. */
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		titleExists = true
	})

	/* Start the collector. */
	c.Visit(searchURL)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}

	return titleExists
}
