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
			scrapeMovieImages := func(title *types.Title) {
				GetTitleImages(title, flags)
			}

			if flags.Async {
				wg.Add(1)
				go func(t *types.Title) {
					defer wg.Done()
					scrapeMovieImages(title)
				}(title)
			} else {
				scrapeMovieImages(title)
			}
			continue // Go to the next title.
		}

		/* For each episode... */
		for _, episode := range title.Episodes {
			/* Get the episode's images. */
			scrapeImages := func(title *types.Title, episode *types.Episode) {
				switch title.Category {
				case types.CategoryAnime, types.CategoryTV:
					GetEpisodeImages(episode, title, flags)
				default:
					fmt.Fprintf(os.Stderr, "Unknown Category: %s (%s) -> [%s]\n", title.Name, title.Link, title.Category)
				}
			}

			if flags.Async {
				wg.Add(1)
				go func(t *types.Title, e *types.Episode) {
					defer wg.Done()
					scrapeImages(t, e)
				}(title, episode)
			} else {
				scrapeImages(title, episode)
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
			fmt.Printf("%s [%s] -> %d images\n", title.Name, title.Category, title.Images.GetImgCount())

			if title.Category == types.CategoryMovie {
				continue // Don't show movie episodes. They don't have any.
			}

			for _, episode := range title.Episodes {
				fmt.Printf("\t%s -> %d images\n", episode.Name, episode.Images.GetImgCount())
			}
		}
	}
}

/*
Given a title `title`, collect its list of images as URLs.

Intended to be used only alongside titles with *NO* episodes. (e.g., Movies)

`title` will have its URL list and image count updated directly from the Title struct.
See `GetEpisodeImages()` for more details on how to handle image collection for titles
with episodes.
*/
func GetTitleImages(title *types.Title, flags cli.CLIFlags) {
	/* Base options for the scraper. */
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains("fancaps.net"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	}

	/* Enable asynchronous mode. */
	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(scraperOpts...)

	c.OnRequest(func(req *colly.Request) {
		req.Headers.Set("Referer", "https://fancaps.net")
	})

	/* Extract a title's images. */
	c.OnHTML("div.row img.imageFade", func(e *colly.HTMLElement) {
		/* Skip "Top Images" */
		if e.DOM.ParentsFiltered("div.topImages").Length() > 0 {
			return
		}

		/* Get image URL. */
		src := e.Attr("src")
		file := path.Base(src)
		imgURL := CategoryURLMap[title.Category] + file

		/* Add URL to list and update image count. */
		title.Images.AddURL(imgURL)
		title.Images.IncrementImgCount()

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

	/* Start the collector on the title link. */
	c.Visit(title.Link)

	/* Wait until all asynchronous requests are complete. */
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
func GetEpisodeImages(episode *types.Episode, title *types.Title, flags cli.CLIFlags) {
	/* Base options for the scraper. */
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains("fancaps.net"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	}

	/* Enable asynchronous mode. */
	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(scraperOpts...)

	c.OnRequest(func(req *colly.Request) {
		req.Headers.Set("Referer", "https://fancaps.net")
	})

	/* Extract an episode's images. */
	c.OnHTML("div.row img.imageFade", func(e *colly.HTMLElement) {
		/* Skip "Top Images" */
		if e.DOM.ParentsFiltered("div.topImages").Length() > 0 {
			return
		}

		/* Get image URL. */
		src := e.Attr("src")
		file := path.Base(src)
		imgURL := CategoryURLMap[title.Category] + file

		/* Add URL to list and update image count. */
		episode.Images.AddURL(imgURL)
		episode.Images.IncrementImgCount()
		title.Images.IncrementImgCount()

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

	/* Start the collector on the title link. */
	c.Visit(episode.Link)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}
}

/* Given a TV Series episode `episode`, return its list of images as URLs. */
func GetTVImages(link string, flags cli.CLIFlags) {
	/* Base options for the scraper. */
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains("fancaps.net"),
	}

	/* Enable asynchronous mode. */
	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(scraperOpts...)

	/* Extract the episode's name and link. (TV-only) */
	c.OnHTML("h3 > a[href]", func(e *colly.HTMLElement) {
		// image := e.Request.AbsoluteURL(e.Attr("href"))
	})

	/*
		If there is a next page,
		visit it to re-trigger episode name/link extraction. (TV-only)
	*/
	c.OnHTML("ul.pager > li > a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPageURL != "#" && containsNext(e.Text) {
			c.Visit(nextPageURL)
		}
	})

	/* Suppress scraper output. */
	if flags.Debug {
		c.OnRequest(func(req *colly.Request) {
			fmt.Println("Visiting TV Episode URL:", req.URL.String())
		})
	}

	/* Start the collector on the title. */
	c.Visit(link)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}
}

/* Given a Movie title `title`, return its list of images as URLs. */
func GetMovieImages(link string, flags cli.CLIFlags) {
	/* Base options for the scraper. */
	scraperOpts := []func(*colly.Collector){
		colly.AllowedDomains("fancaps.net"),
	}

	/* Enable asynchronous mode. */
	if flags.Async {
		scraperOpts = append(scraperOpts, colly.Async(true))
	}

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(scraperOpts...)

	/* Extract the episode's name and link. (TV-only) */
	c.OnHTML("h3 > a[href]", func(e *colly.HTMLElement) {
		// image := e.Request.AbsoluteURL(e.Attr("href"))
	})

	/*
		If there is a next page,
		visit it to re-trigger episode name/link extraction. (TV-only)
	*/
	c.OnHTML("ul.pager > li > a[href]", func(e *colly.HTMLElement) {
		nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
		if nextPageURL != "#" && containsNext(e.Text) {
			c.Visit(nextPageURL)
		}
	})

	/* Suppress scraper output. */
	if flags.Debug {
		c.OnRequest(func(req *colly.Request) {
			fmt.Println("Visiting TV Episode URL:", req.URL.String())
		})
	}

	/* Start the collector on the title. */
	c.Visit(link)

	/* Wait until all asynchronous requests are complete. */
	if flags.Async {
		c.Wait()
	}
}
