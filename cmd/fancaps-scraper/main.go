package main

import (
	"fmt"
	"os"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/menu"
	"sheeper.com/fancaps-scraper-go/pkg/scraper"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

func main() {
	flags := cli.ParseCLI()

	searchURL := flags.BuildQueryURL()
	titles := scraper.GetTitles(searchURL)

	/* Debug: Print found titles. */
	fmt.Println("FOUND TITLES:")
	for _, t := range titles {
		fmt.Println(t.Name, t.Link)
	}
	fmt.Println()

	/* Get the episodes for each title. */
	for i := range titles {
		/*
			From title category, run corresponding episode scraper.
			Note: Movies do not have episodes and thus do not require episode scraping.
		*/
		switch titles[i].Category {
		case types.CategoryAnime:
			titles[i].Episodes = scraper.GetAnimeEpisodes(titles[i])
		case types.CategoryTV:
			titles[i].Episodes = scraper.GetTVEpisodes(titles[i])
		case types.CategoryMovie:
			// Do nothing
		default:
			fmt.Fprintf(os.Stderr, "Unknown Category: %s (%s) -> [%s]\n", titles[i].Name, titles[i].Link, titles[i].Category)
		}
	}

	/* Debug: Print found titles and episodes. */
	fmt.Println("\nFOUND TITLES AND EPISODES:")
	for _, title := range titles {
		fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
		for _, episode := range title.Episodes {
			fmt.Printf("\t%s -> %s\n", episode.Name, episode.Link)
		}
	}

	selectedTitles, confirmed := menu.GetTitleMenu(titles)
	if !confirmed {
		fmt.Fprintf(os.Stderr, "Title Menu: Operation aborted.\n")
		os.Exit(1)
	}

	/* Debug: Print selected titles and episodes. */
	fmt.Println("\nSELECTED TITLES AND EPISODES:")
	for _, title := range selectedTitles {
		fmt.Printf("%s [%s] -> %s\n", title.Name, title.Category, title.Link)
		for _, episode := range title.Episodes {
			fmt.Printf("\t%s -> %s\n", episode.Name, episode.Link)
		}
	}
}
