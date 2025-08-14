package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	enumflag "sheeper.com/fancaps-scraper-go/pkg/cli/custom/enum"
	fsflag "sheeper.com/fancaps-scraper-go/pkg/cli/custom/fs"
	numflag "sheeper.com/fancaps-scraper-go/pkg/cli/custom/number"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Available CLI Flags. */
type CLIFlags struct {
	Queries           []string         // Search queries to scrape from.
	Categories        []types.Category // Selected categories to search using `Query`.
	OutputDir         string           // The directory to output images.
	ParallelDownloads uint8            // Maximum amount of image downloads to make in parallel.
	MinDelay          time.Duration    // Minimum delay applied after subsequent image requests. (Non-negative)
	RandDelay         time.Duration    // Maximum random delay applied after subsequent image requests. (Non-negative)
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

	defaultParallelDownloads uint8         = 10              // Default maximum amount of titles or episodes to download images from in parallel.
	defaultMinDelay          time.Duration = 1 * time.Second // Default minimum delay after every new image download request.
	defaultRandDelay         time.Duration = 5 * time.Second // Default maximum random delay after every new image download request.
)

var (
	defaultCategories = []types.Category{
		types.CategoryAnime,
		types.CategoryTV,
		types.CategoryMovie,
	} // Default categories to search.
	enumToCategory = map[string]types.Category{
		"anime":  types.CategoryAnime,
		"tv":     types.CategoryTV,
		"movies": types.CategoryMovie,
	} // A map from custom enums to categories.
)

var defaultOutputDir = filepath.Join(".", "output") // Default output directory.

/* Parses CLI flags. */
func ParseCLI() {
	var (
		queries           []string
		categories        []types.Category
		outputDir         string
		parallelDownloads uint8
		minDelay          time.Duration
		randDelay         time.Duration
		async             bool
		debug             bool
		nolog             bool
	)

	rootCmd := &cobra.Command{
		Use:     "fancaps-scraper",
		Short:   "Scrape images from fancaps.net using a CLI interface",
		Example: exampleUsage,
		Run: func(cmd *cobra.Command, args []string) {
			flags.Queries = queries
			flags.Categories = categories
			flags.OutputDir = outputDir
			flags.ParallelDownloads = parallelDownloads
			flags.MinDelay = minDelay
			flags.RandDelay = randDelay
			flags.Async = async
			flags.Debug = debug
			flags.NoLog = nolog
		},
	}

	/* Flag Definitions. */
	rootCmd.Flags().StringSliceVarP(&queries, "query", "q", []string{}, "Search query terms.")
	enumflag.EnumSliceVarP(rootCmd.Flags(), &categories, "categories", "c", defaultCategories, enumToCategory, "Categories to search.")
	fsflag.CreateDirVarP(rootCmd.Flags(), &outputDir, "output-dir", "o", defaultOutputDir, "Output directory for images.")
	numflag.Puint8VarP(rootCmd.Flags(), &parallelDownloads, "parallel-downloads", "p", defaultParallelDownloads, "Maximum amount of image downloads to request in parallel.")
	numflag.NnDurationVar(rootCmd.Flags(), &minDelay, "min-delay", defaultMinDelay, "Minimum delay applied after subsequent image requests.")
	numflag.NnDurationVar(rootCmd.Flags(), &randDelay, "random-delay", defaultRandDelay, "Maximum random delay applied after subsequent image requests.")
	rootCmd.Flags().BoolVar(&async, "async", true, "Enable asynchronous requests.")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode.")
	rootCmd.Flags().BoolVar(&nolog, "no-log", false, "Disable logging.")

	/* "Override" default help. */
	rootCmd.Flags().BoolP("help", "h", false, "Display this help and exit.")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Root().Usage()
		os.Exit(0)
	})

	/* Show flags in the order they were defined. */
	rootCmd.Flags().SortFlags = false

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/* Returns a copy of the CLI flags. */
func Flags() CLIFlags {
	return flags
}
