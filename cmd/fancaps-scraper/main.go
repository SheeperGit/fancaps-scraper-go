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

	episodes := scraper.GetEpisodes(titles)

	/* Debug: Print found episodes. */
	fmt.Println("Found Episodes:")
	for _, e := range episodes {
		fmt.Printf("%s -> %s\n", e.Name, e.Link)
	}
}
