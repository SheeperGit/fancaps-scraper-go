package main

import (
	"fmt"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
)

func main() {
	flags := cli.ParseCLI()

	searchURL := flags.BuildQueryURL()
	titles := scraper.GetTitles(searchURL)

	/* Debug: Print found titles. */
	fmt.Println("Found Titles:")
	for _, t := range titles {
		fmt.Println(t.Name, t.Link)
	}

	/* Get the episodes for each title. */
	for i := range titles {
		titles[i].Episodes = titles[i].GetEpisodes()
	}

	/* Debug: Print found titles and episodes. */
	fmt.Println("FULL INFO:")
	for _, title := range titles {
		fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
		for _, episode := range title.Episodes {
			fmt.Printf("\t%s -> %s\n", episode.Name, episode.Link)
		}
	}
}
