package main

import (
	"fmt"
	"os"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/menu"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
)

func main() {
	flags := cli.ParseCLI()

	if !flags.Movies && !flags.TV && !flags.Anime {
		selectedMenuCategories, confirmed := menu.GetCategoriesMenu()
		if !confirmed {
			fmt.Fprintf(os.Stderr, "Category Menu: Operation aborted.\n")
			os.Exit(1)
		}

		/* Set active categories according to Category Menu. */
		for cat := range selectedMenuCategories {
			switch cat {
			case menu.MOVIE_TEXT:
				flags.Movies = true
			case menu.TV_TEXT:
				flags.TV = true
			case menu.ANIME_TEXT:
				flags.Anime = true
			}
		}
	}

	searchURL := flags.BuildQueryURL()
	titles := scraper.GetTitles(searchURL)

	/* Debug: Print found titles. */
	fmt.Println("Found Titles:")
	for _, t := range titles {
		fmt.Println(t.Name, t.Link)
	}
}
