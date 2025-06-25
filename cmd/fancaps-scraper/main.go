package main

import (
	"fmt"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/menu"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
)

func main() {
	/* Parse flags. */
	flags := cli.ParseCLI()

	/* Get the URL to scrape based on category selections. */
	searchURL := flags.BuildQueryURL()

	/* Get titles matching user query. */
	titles := scraper.GetTitles(searchURL, flags)

	/* Get episodes from titles. */
	scraper.GetEpisodes(titles, flags)

	/* Allow the user to choose which titles and episodes to scrape from. */
	selectedTitles := menu.LaunchTitleMenu(titles, flags.Categories, flags.Debug)

	// to stop the compiler from complaining that the variable is unused
	if len(selectedTitles) != 0 {
		fmt.Println("yup, we sure got our selected titles :)")
	}
}
