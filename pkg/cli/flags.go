package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"sheeper.com/fancaps-scraper-go/pkg/fsutil"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
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
	NoLog             bool             // If true, disable logging.
}

var flags CLIFlags // User CLI flags.

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

/* Parses CLI flags. */
func ParseCLI() {
	var (
		queries           []string
		categories        string
		outputDir         string
		parallelDownloads uint8
		minDelay          uint32
		randDelay         uint32
		async             bool
		debug             bool
		nolog             bool
	)

	rootCmd := &cobra.Command{
		Use:     "fancaps-scraper",
		Short:   "Scrape images from fancaps.net using a CLI interface",
		Example: exampleUsage,
		Run: func(cmd *cobra.Command, args []string) {
			/* Check that the parent directories exist. */
			if !fsutil.ParentDirsExist(outputDir) {
				fmt.Fprintf(os.Stderr,
					ui.ErrStyle.Render("couldn't find parent directories of `%s`")+"\n"+
						ui.ErrStyle.Render("make sure the parent directories exists.")+"\n",
					outputDir)
				os.Exit(1)
			}
			flags.OutputDir = outputDir

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

			flags.Queries = queries
			flags.MinDelay = minDelay
			flags.RandDelay = randDelay
			flags.Async = async
			flags.Debug = debug
			flags.NoLog = nolog
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
	rootCmd.Flags().BoolVar(&nolog, "no-log", false, "Disable logging")

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
}

/* Returns a copy of the CLI flags. */
func Flags() CLIFlags {
	return flags
}
