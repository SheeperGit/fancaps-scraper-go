package main

import (
	"fmt"
	"os"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/menu"
)

func main() {
	flags := cli.ParseCLI()

	if !flags.Movies && !flags.TV && !flags.Anime {
		selectedMenuCategories, confirmed := menu.GetCategoryMenu()
		if !confirmed {
			fmt.Printf("Category Menu: Operation aborted.\n")
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

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
		colly.Async(true),
	)

	/*
		On every h4 element which has an anchor child element,
		print the link text and the link itself.
	*/
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	})

	/* Before making a request, print "Visiting ..." */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector. */
	c.Visit(searchURL)

	/* Wait for the collector to finish. (Required for Async) */
	c.Wait()
}
