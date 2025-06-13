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

	searchURL := flags.BuildQueryURL()
	titles := scraper.GetTitles(searchURL)

	/* Debug: Print found titles. */
	fmt.Println("Found Titles:")
	for _, t := range titles {
		fmt.Println(t.Name, t.Link)
	}

	/* Get the episodes for each title. */
	for i := range titles {
		/*
			From title category, run corresponding episode scraper.
			Note: Movies do not have episodes and thus do not require episode scraping.
		*/
		switch titles[i].Category {
		case menu.CategoryAnime:
			titles[i].Episodes = titles[i].GetAnimeEpisodes()
		case menu.CategoryTV:
			titles[i].Episodes = titles[i].GetTVEpisodes()
		case menu.CategoryMovie:
			// Do nothing
		default:
			fmt.Fprintf(os.Stderr, "Unknown Category: %s (%s) -> [%s]\n", titles[i].Name, titles[i].Link, titles[i].Category)
		}
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
