package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"sheeper.com/fancaps-scraper-go/pkg/fsutil"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
)

/*
Returns the output directory `outputDir` if its parent directories exist,
and exits with status code 1 otherwise.
*/
func validateOutputDir(outputDir string) string {
	if !fsutil.ParentDirsExist(outputDir) {
		fmt.Fprintf(os.Stderr,
			ui.ErrStyle.Render("couldn't find parent directories of `%s`")+"\n"+
				ui.ErrStyle.Render("make sure the parent directories exists.")+"\n",
			outputDir)
		os.Exit(1)
	}

	return outputDir
}

/*
Returns the amount of parallel downloads to make if it is greater than 0,
and exits with status code 1 otherwise.
*/
func validateParallelDownloads(parallelDownloads uint8) uint8 {
	if parallelDownloads == 0 {
		fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("parallel downloads must be stricly positive."))
		os.Exit(1)
	}

	return parallelDownloads
}

/*
Returns a non-empty sorted, parsed list of categories from categories `categories` if non-empty,
and prompts the user for categories with a menu otherwise.

If `categories` contains an unknown category, this function exists with status code 1.
*/
func parseCategories(categories []string) []types.Category {
	cats := []types.Category{}

	if len(categories) != 0 { // Parse provided categories.
		categoryMap := map[string]types.Category{
			"anime":  types.CategoryAnime,
			"tv":     types.CategoryTV,
			"movies": types.CategoryMovie,
		}
		seen := map[types.Category]bool{}

		for _, part := range categories {
			part = strings.ToLower(part)
			if part == "all" {
				for _, cat := range categoryMap {
					if !seen[cat] {
						cats = append(cats, cat)
						seen[cat] = true
					}
				}
				break
			}

			if cat, ok := categoryMap[part]; ok && !seen[cat] {
				cats = append(cats, cat)
				seen[cat] = true
			} else if !ok {
				fmt.Fprintf(os.Stderr, "unknown category `%s`. valid options are: anime, tv, movies, all\n", part)
				os.Exit(1)
			}
		}
	} else { // No categories specified, prompt user for categories with a Category Menu.
		selectedMenuCategories := menu.LaunchCategoriesMenu()
		for cat := range selectedMenuCategories {
			cats = append(cats, cat)
		}
	}

	/* Sort according to Category enum order. */
	slices.Sort(cats)

	return cats
}
