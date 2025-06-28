package main

import (
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/menu"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

func main() {
	/* Parse flags. */
	flags := cli.ParseCLI()

	/* Get the URL to scrape based on category selections. */
	searchURL := flags.BuildQueryURL()

	/* Category statistics. */
	catStats := types.NewCatStats()

	/* Get titles matching user query. */
	titles := scraper.GetTitles(searchURL, catStats, flags)

	/* Allow the user to choose which titles to scrape from. */
	selectedTitles := menu.LaunchTitleMenu(titles, flags.Categories, catStats, flags.Debug)

	/* Get episodes from selected titles. */
	scraper.GetEpisodes(selectedTitles, flags)
}
