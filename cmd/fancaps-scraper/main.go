package main

import (
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/logf"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
	"sheeper.com/fancaps-scraper-go/pkg/ui/prompt"
)

func main() {
	/* Get search URLs and parse flags. */
	searchURLs, flags := cli.ParseCLI()

	/* Get titles matching user query. */
	titles := scraper.GetTitles(searchURLs)

	/* Allow the user to choose which titles to scrape from. */
	selectedTitles := menu.LaunchTitleMenu(titles, flags.Categories, flags.Debug)

	/* Get episodes from selected titles. */
	scraper.GetEpisodes(selectedTitles)

	/* Select episodes to scrape from each title. */
	prompt.SelectEpisodes(selectedTitles, flags.Debug)

	/* Collect images from the selected titles and episodes. */
	scraper.GetImages(selectedTitles)

	/* Download images from the selected titles and episodes. */
	scraper.DownloadImages(selectedTitles)

	/* Print info that may require user attention. Otherwise, indicate success. */
	logf.PrintStats()
}
