package cli

import (
	"path/filepath"
	"time"

	"sheeper.com/fancaps-scraper-go/pkg/types"
)

const (
	exampleUsage = `Usage:
	fancaps-scraper-go [OPTIONS]

Examples:
	# Show this message and exit.
  fancaps-scraper --help

  # Search for "Naruto" with anime and tv series titles only.
  fancaps-scraper --query Naruto --categories anime,tv

  # Search for "The Office" (with short flags). (Notice also the single quotes to signify "The Office" as one argument.)
  fancaps-scraper -q 'The Office'

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

	defaultOutputDir = filepath.Join(".", "output") // Default output directory.
)
