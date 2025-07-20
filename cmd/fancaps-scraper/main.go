package main

import (
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
	"sheeper.com/fancaps-scraper-go/pkg/ui/prompt"
)

func main() {
	/* Parse flags. */
	flags := cli.ParseCLI()

	/* Get the URL to scrape based on category selections. */
	searchURL := cli.BuildQueryURL(flags.Query, flags.Categories)

	/* Category statistics. */
	catStats := types.NewCatStats()

	/* Get titles matching user query. */
	titles := scraper.GetTitles(searchURL, catStats, flags)

	/* Allow the user to choose which titles to scrape from. */
	selectedTitles := menu.LaunchTitleMenu(titles, flags.Categories, catStats, flags.Debug)

	/* Get episodes from selected titles. */
	scraper.GetEpisodes(selectedTitles, flags)

	/* Select episodes to scrape from each title. */
	prompt.SelectEpisodes(selectedTitles, flags.Debug)

	/* Collect images from the selected titles and episodes. */
	scraper.GetImages(selectedTitles, flags)

	/* Download images from the selected titles and episodes. */
	scraper.DownloadImages(selectedTitles, flags)
}
