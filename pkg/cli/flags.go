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
	"golang.org/x/sync/errgroup"
	"sheeper.com/fancaps-scraper-go/pkg/logf"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
	"sheeper.com/fancaps-scraper-go/pkg/ui/prompt"
)

/* Available CLI Flags. */
type CLIFlags struct {
	Queries           []string         // Search queries to scrape from.
	Categories        []types.Category // Selected categories to search using `Query`.
	OutputDir         string           // The directory to output images.
	ParallelDownloads uint8            // Maximum amount of image downloads to make in parallel.
	MinDelay          uint32           // Minimum delay applied after subsequent image requests. (In milliseconds)
	RandDelay         uint32           // Maximum random delay applied after subsequent image requests. (In milliseconds)
	Async             bool             // If true, enable asynchronous network requests.
	Debug             bool             // If true, print useful debugging messages.
}

const (
	exampleUsage = `
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

	defaultParallelDownloads uint8  = 10   // Default maximum amount of titles or episodes to download images from in parallel.
	defaultMinDelay          uint32 = 1000 // Default minimum delay (in milliseconds) after every new image download request.
	defaultRandDelay         uint32 = 5000 // Default maximum random delay (in milliseconds) after every new image download request.
)

var defaultOutputDir = filepath.Join(".", "output") // Default output directory.

/*
Returns a list of search URLs `searchURLs` to scrape from and the parsed CLI flags.

`flags.Queries` always contains a non-empty list of queries with at least one title
associated with each query's search URL.
*/
func ParseCLI() ([]string, CLIFlags) {
	var (
		flags             CLIFlags
		queries           []string
		categories        string
		outputDir         string
		parallelDownloads uint8
		minDelay          uint32
		randDelay         uint32
		async             bool
		debug             bool
	)

	var searchURLs []string

	rootCmd := &cobra.Command{
		Use:     "fancaps-scraper",
		Short:   "Scrape images from fancaps.net using a CLI interface",
		Example: exampleUsage,
		Run: func(cmd *cobra.Command, args []string) {
			/* Check that the parent directories exist. */
			if !ParentDirsExist(outputDir) {
				fmt.Fprintf(os.Stderr,
					ui.ErrStyle.Render("couldn't find parent directories of `%s`")+"\n"+
						ui.ErrStyle.Render("make sure the parent directories exists.")+"\n",
					outputDir)
				os.Exit(1)
			}
			flags.OutputDir = outputDir // Title directories go here.
			logf.LogDir = outputDir     // Log file goes here too.

			/* Check that parallel downloads is non-zero. */
			if parallelDownloads == 0 {
				fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("parallel downloads must be set stricly positive."))
				os.Exit(1)
			}
			flags.ParallelDownloads = parallelDownloads

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
						fmt.Fprintf(os.Stderr, "unknown category `%s`. valid options are: anime, tv, movies, all\n", part)
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

			flags.Async = async

			/* Query validation. */
			if !cmd.Flags().Changed("query") {
				/*
					If `-q` not specified, prompt user for search query.
					Validate search URLs incrementally.
				*/
				for len(queries) == 0 || prompt.YesNoPrompt("Enter another query? [y/N]: ", "") {
					query := prompt.TextPrompt("Enter Search Query: ", prompt.QueryHelpPrompt)
					if strings.TrimSpace(query) == "" {
						fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("search query cannot be empty.\n\n"))
						continue
					}

					url := BuildQueryURL(query, flags.Categories)
					if !titleExists(url, flags) {
						fmt.Fprintf(os.Stderr, ui.ErrStyle.Render("no titles found for query `%s`.")+"\n\n", query)
						continue
					}
					fmt.Printf(ui.SuccessStyle.Render("Found titles for query: `%s`")+"\n", query)
					searchURLs = append(searchURLs, url)
					queries = append(queries, query)
				}
				flags.Queries = queries
			} else {
				/* Validate search URLs all at once. */
				for _, query := range queries {
					if strings.TrimSpace(query) == "" {
						fmt.Fprintln(os.Stderr, "search query cannot be empty.")
						os.Exit(1)
					}
					url := BuildQueryURL(query, flags.Categories)
					searchURLs = append(searchURLs, url)
				}

				var eg errgroup.Group
				for i, url := range searchURLs {
					i, url := i, url // https://golang.org/doc/faq#closures_and_goroutines
					eg.Go(func() error {
						if !titleExists(url, flags) {
							return fmt.Errorf("no titles found for query `%s`", queries[i])
						}
						return nil
					})

					if err := eg.Wait(); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
			}

			flags.MinDelay = minDelay
			flags.RandDelay = randDelay
			flags.Debug = debug
		},
	}

	/* Flag Definitions. */
	rootCmd.Flags().StringSliceVarP(&queries, "query", "q", []string{}, "Search query terms")
	rootCmd.Flags().StringVarP(&categories, "categories", "c", "", "Categories to search. Format: [anime,tv,movies|all] (comma-separated)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", defaultOutputDir, "Output directory for images. (Parent directories must exist)")
	rootCmd.Flags().Uint8VarP(&parallelDownloads, "parallel-downloads", "p", defaultParallelDownloads, "Maximum amount of image downloads to request in parallel.")
	rootCmd.Flags().Uint32Var(&minDelay, "min-delay", defaultMinDelay, "Minimum delay applied after subsequent image requests. (In milliseconds)")
	rootCmd.Flags().Uint32Var(&randDelay, "random-delay", defaultRandDelay, "Maximum random delay applied after subsequent image requests. (In milliseconds)")
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

	return searchURLs, flags
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
