package scraper

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* Base URLs from which the image files are stored. */
const (
	baseAnimeURL = "https://cdni.fancaps.net/file/fancaps-animeimages/"
	baseTVURL    = "https://cdni.fancaps.net/file/fancaps-tvimages/"
	baseMovieURL = "https://cdni.fancaps.net/file/fancaps-movieimages/"
)

/* Maps a category to its base URL where its image URLs are stored. */
var CategoryURLMap = map[types.Category]string{
	types.CategoryAnime: baseAnimeURL,
	types.CategoryTV:    baseTVURL,
	types.CategoryMovie: baseMovieURL,
}

/* Get images from titles `titles`. */
func GetImages(titles []*types.Title, flags cli.CLIFlags) {
	var wg sync.WaitGroup

	/* For each title... */
	for _, title := range titles {
		/* Handle movies seperately, since they have no episodes. */
		if title.Category == types.CategoryMovie {
			if flags.Async {
				wg.Add(1)
				go func(t *types.Title) {
					defer wg.Done()
					scrapeTitleImages(title, flags)
				}(title)
			} else {
				scrapeTitleImages(title, flags)
			}
			continue // Go to the next title.
		}

		/* For each episode... */
		for _, episode := range title.Episodes {
			/* Get the episode's images. */
			scrapeEpisodeImgs := func(title *types.Title, episode *types.Episode) {
				switch title.Category {
				case types.CategoryAnime, types.CategoryTV:
					scrapeEpisodeImages(episode, title, flags)
				default:
					fmt.Fprintf(os.Stderr, "Unknown Category: %s (%s) -> [%s]\n", title.Name, title.Link, title.Category)
				}
			}

			if flags.Async {
				wg.Add(1)
				go func(t *types.Title, e *types.Episode) {
					defer wg.Done()
					scrapeEpisodeImgs(t, e)
				}(title, episode)
			} else {
				scrapeEpisodeImgs(title, episode)
			}
		}
	}

	if flags.Async {
		wg.Wait()
	}

	/* Debug: Print amount of found images per title/episode. */
	if flags.Debug {
		fmt.Println("\n\nFOUND IMAGES:")
		for _, title := range titles {
			fmt.Printf("%s [%s] -> %d images\n", title.Name, title.Category, title.Images.Total())

			if title.Category == types.CategoryMovie {
				continue // Don't show movie episodes. They don't have any.
			}

			for _, episode := range title.Episodes {
				fmt.Printf("\t%s -> %d images\n", episode.Name, episode.Images.Total())
			}
		}
		fmt.Printf("\n\n")
	}
}

/*
Given a title `title`, collect its list of images as URLs.

Intended to be used only alongside titles with *NO* episodes. (e.g., Movies)

`title` will have its URL list and image count updated directly from the Title struct.
See `GetEpisodeImages()` for more details on how to handle image collection for titles
with episodes.
*/
func scrapeTitleImages(title *types.Title, flags cli.CLIFlags) {
	scraperOpts := GetScraperOpts(flags)
	c := colly.NewCollector(scraperOpts...)

	/* Extract title image. */
	c.OnHTML("div.row img.imageFade", func(e *colly.HTMLElement) {
		/* Skip "Top Images". (They will be downloaded anyway.) */
		if e.DOM.ParentsFiltered("div.topImages").Length() > 0 {
			return
		}

		/* Get image URL. */
		src := e.Attr("src")
		file := path.Base(src)
		imgURL := CategoryURLMap[title.Category] + file

		title.Images.AddURL(imgURL)
		title.IncrementImageTotal()

		if flags.Debug {
			fmt.Printf("%s [%s] image found! (%s)\n", title.Name, title.Category, imgURL)
		}
	})

	/*
		If there is a next page,
		visit it to re-trigger episode image extraction. (Anime-only)
	*/
	c.OnHTML("ul.pagination > li > a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if e.Text == "»" && nextPageURL != "#" {
			c.Visit(nextPageURL)
		}
	})

	c.Visit(title.Link)

	if flags.Async {
		c.Wait()
	}
}

/*
Given an episode `episode`, collect its list of images as URLs
and set the state of its title `title`, accordingly.

Intended to be used only alongside titles with episodes. (e.g., Anime, TV Series)

`title` will only have its image count updated, its URL list will
be left alone. This is intentional, as only Movie titles will directly store all
of their URLs in the Title struct. See `GetTitleImages()` for more details.
*/
func scrapeEpisodeImages(episode *types.Episode, title *types.Title, flags cli.CLIFlags) {
	scraperOpts := GetScraperOpts(flags)
	c := colly.NewCollector(scraperOpts...)

	/* Extract episode image. */
	c.OnHTML("div.row img.imageFade", func(e *colly.HTMLElement) {
		/* Skip "Top Images". (They will be downloaded anyway.) */
		if e.DOM.ParentsFiltered("div.topImages").Length() > 0 {
			return
		}

		/* Get image URL. */
		src := e.Attr("src")
		file := path.Base(src)
		imgURL := CategoryURLMap[title.Category] + file

		episode.Images.AddURL(imgURL)
		episode.IncrementImageTotal()

		if flags.Debug {
			fmt.Printf("%s - %s [%s] image found! (%s)\n", title.Name, episode.Name, title.Category, imgURL)
		}
	})

	/*
		If there is a next page,
		visit it to re-trigger episode image extraction. (Anime-only)
	*/
	c.OnHTML("ul.pagination > li > a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if e.Text == "»" && nextPageURL != "#" {
			c.Visit(nextPageURL)
		}
	})

	c.Visit(episode.Link)

	if flags.Async {
		c.Wait()
	}
}
