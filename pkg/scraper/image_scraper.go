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
var CategoryName = map[types.Category]string{
	types.CategoryAnime: baseAnimeURL,
	types.CategoryTV:    baseTVURL,
	types.CategoryMovie: baseMovieURL,
}

/* Get images from titles `titles`. */
func GetImages(titles []types.Title, flags cli.CLIFlags) {
	var wg sync.WaitGroup

	/* For each title... */
	for i := range titles {
		/* Handle movies seperately, since they have no episodes. */
		if titles[i].Category == types.CategoryMovie {
			scrapeMovieImages := func(i int) {
				// GetMovieImages(titles[i].Link, flags)
			}

			if flags.Async {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					scrapeMovieImages(i)
				}(i)
			} else {
				scrapeMovieImages(i)
			}
			continue // Go to the next title.
		}

		/* For each episode... */
		for j := range titles[i].Episodes {
			/* Get the episode's images. */
			scrapeImages := func(i, j int) {
				switch titles[i].Category {
				case types.CategoryAnime:
					GetAnimeImages(&titles[i], titles[i].Episodes[j], flags)
				case types.CategoryTV:
					// GetTVImages(titles[i].Episodes[j].Link, flags)
				default:
					fmt.Fprintf(os.Stderr, "Unknown Category: %s (%s) -> [%s]\n", titles[i].Name, titles[i].Link, titles[i].Category)
				}
			}

			if flags.Async {
				wg.Add(1)
				go func(i, j int) {
					defer wg.Done()
					scrapeImages(i, j)
				}(i, j)
			} else {
				scrapeImages(i, j)
			}
		}
	}

	if flags.Async {
		wg.Wait()
	}

	/* Debug: Print amount of found images per title/episode. */
	if flags.Debug {
		fmt.Println("\n\nFOUND IMAGES:")
		for i := range titles {
			title := &titles[i]
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

/* Given an Anime episode `episode`, return its list of images as URLs. */
func GetAnimeImages(title *types.Title, episode *types.Episode, flags cli.CLIFlags) {
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

	/* Extract an episode's images. (Anime-only) */
	c.OnHTML("div.row img.imageFade", func(e *colly.HTMLElement) {
		/* Skip "Top Images" */
		if e.DOM.ParentsFiltered("div.topImages").Length() > 0 {
			return
		}

		/* Save Image. */
		src := e.Attr("src")
		file := path.Base(src)
		imgURL := baseAnimeURL + file

		if title.Category != types.CategoryMovie {
			episode.Images.AddURL(imgURL)
			episode.Images.IncrementImgCount()
		} else {
			title.Images.AddURL(imgURL)
		}

		title.Images.IncrementImgCount()
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
