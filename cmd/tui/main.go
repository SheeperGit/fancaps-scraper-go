package main

import (
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/format"
	"sheeper.com/fancaps-scraper-go/pkg/logf"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
	"sheeper.com/fancaps-scraper-go/pkg/ui/menu"
	"sheeper.com/fancaps-scraper-go/pkg/ui/prompt"
)

func main() {
	/* Parse flags. */
	cli.ParseCLI()

	/* Get parsed flags. */
	flags := cli.Flags()

	/* Get URLs to search through. */
	searchURLs := scraper.GetSearchURLs(flags.Queries, flags.Categories)

	/* Get titles matching user query. */
	titles := scraper.GetTitles(searchURLs)

	/* Allow the user to choose which titles to scrape from. */
	selectedTitles := menu.LaunchTitleMenu(titles, flags.Categories, flags.MenuLines, flags.Debug)

	/* Get episodes from selected titles. */
	scraper.GetEpisodes(selectedTitles)

	/* Select episodes to scrape from each title. */
	prompt.SelectEpisodes(selectedTitles, flags.Debug)

	/* Collect images from the selected titles and episodes. */
	scraper.GetImages(selectedTitles)

	if flags.DryRun { /* Dry run mode: Print data, don't download anything. */
		format.OutputFormat(selectedTitles, flags.Format.String())
	} else { /* Download images from the selected titles and episodes. */
		scraper.DownloadImages(selectedTitles)
	}

	/* Print info that may require user attention. Otherwise, indicate success. */
	logf.PrintStats()
}
