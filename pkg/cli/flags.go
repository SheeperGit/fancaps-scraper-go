package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	enumflag "sheeper.com/fancaps-scraper-go/pkg/cli/custom/enum"
	fsflag "sheeper.com/fancaps-scraper-go/pkg/cli/custom/fs"
	numflag "sheeper.com/fancaps-scraper-go/pkg/cli/custom/number"
	"sheeper.com/fancaps-scraper-go/pkg/format"
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
	Verbose           bool             // If true, explain what is being done.
	Debug             bool             // If true, print useful debugging messages.
	NoLog             bool             // If true, disable logging.
	DryRun            bool             // If true, perform a dry run. (Safe. No changes made.)
	Format            format.Format    // Format used to print scraped titles.
}

var flags CLIFlags // User CLI flags.

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
		verbose           bool
		debug             bool
		nolog             bool
		dryRun            bool
		format            format.Format
	)

	f := pflag.NewFlagSet("fancaps-scraper", pflag.ContinueOnError)

	/* Show flags in the order they were defined. */
	f.SortFlags = false

	/* Define how usage is shown. */
	f.Usage = func() {
		if exampleUsage != "" {
			fmt.Fprintln(f.Output(), exampleUsage)
			fmt.Fprintln(f.Output(), "")
		}
		fmt.Fprintln(f.Output(), "Flags:")
		f.PrintDefaults()
	}

	/* Flag Definitions. */
	f.StringSliceVarP(&queries, "query", "q", []string{}, "Search query terms.")
	enumflag.EnumSliceVarP(f, &categories, "categories", "c", defaultCategories, enumToCategory, "Categories to search.")
	fsflag.CreateDirVarP(f, &outputDir, "output-dir", "o", defaultOutputDir, "Output directory for images.")
	numflag.Puint8VarP(f, &parallelDownloads, "parallel-downloads", "p", defaultParallelDownloads, "Maximum concurrent image downloads.")
	numflag.NnDurationVar(f, &minDelay, "min-delay", defaultMinDelay, "Minimum delay between image requests.")
	numflag.NnDurationVar(f, &randDelay, "random-delay", defaultRandDelay, "Maximum random delay between image requests.")
	f.BoolVar(&async, "async", true, "Enable asynchronous requests.")
	f.BoolVarP(&verbose, "verbose", "v", false, "Display what is being done.")
	f.BoolVar(&debug, "debug", false, "Display results as stages complete.")
	f.BoolVar(&nolog, "no-log", false, "Disable logging.")
	f.BoolVarP(&dryRun, "dry-run", "n", false, "Do not change anything, only print results.")
	enumflag.EnumVar(f, &format, "format", defaultFormat, enumToFormat, "Output format for dry-run.")

	/* Custom help. */
	var help bool
	f.BoolVarP(&help, "help", "h", false, "Display this help and exit.")

	/* Parse args. */
	if err := f.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/* Show usage, if requested. */
	if help {
		f.Usage()
		os.Exit(0)
	}

	/* Assign values. */
	flags.Queries = queries
	flags.Categories = categories
	flags.OutputDir = outputDir
	flags.ParallelDownloads = parallelDownloads
	flags.MinDelay = minDelay
	flags.RandDelay = randDelay
	flags.Async = async
	flags.Verbose = verbose
	flags.Debug = debug
	flags.NoLog = nolog
	flags.DryRun = dryRun
	flags.Format = format
}

/* Returns a copy of the CLI flags. */
func Flags() CLIFlags {
	return flags
}
