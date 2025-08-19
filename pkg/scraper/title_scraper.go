package scraper

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

var seasonRegex = regexp.MustCompile(` Season (\d+)`) // Extracts a title's season number.

/*
Returns a non-empty, unique, sorted list of titles found through the URLs `searchURLs`,
which are assumed to be validated (have at least one title associated with each URL).
*/
func GetTitles(searchURLs []string) []*types.Title {
	var (
		titles   []*types.Title              // Scraped titles.
		titlesMu sync.Mutex                  // Prevents overlapping "appends" to `titles`.
		wg       sync.WaitGroup              // Synchronizes title scrapers.
		seen     = make(map[string]struct{}) // Duplicate titles protection.
	)

	flags := cli.Flags()

	for _, searchURL := range searchURLs {
		wg.Add(1)
		go func(searchURL string) {
			defer wg.Done()

			ts := scrapeTitles(searchURL, flags)

			titlesMu.Lock()
			for _, t := range ts {
				if _, exists := seen[t.Url]; !exists {
					seen[t.Url] = struct{}{}
					titles = append(titles, t)
				}
			}
			titlesMu.Unlock()
		}(searchURL)
	}
	wg.Wait()

	/*
		Sort found titles. (Case-insensitive)
		Precedence (Highest to Lowest): Category, Name + Season, Name.
	*/
	sort.Slice(titles, func(i, j int) bool {
		catI := titles[i].Category
		catJ := titles[j].Category

		if catI != catJ {
			return catI < catJ
		}

		nameI := titles[i].Name
		nameJ := titles[j].Name

		baseNameI := strings.ToLower(seasonRegex.ReplaceAllString(nameI, ""))
		baseNameJ := strings.ToLower(seasonRegex.ReplaceAllString(nameJ, ""))

		if baseNameI != baseNameJ {
			return baseNameI < baseNameJ
		}

		numI, seasonFoundI := getSeasonNumber(nameI)
		numJ, seasonFoundJ := getSeasonNumber(nameJ)

		if seasonFoundI && seasonFoundJ {
			return numI < numJ
		}

		return baseNameI < baseNameJ
	})

	/* Debug: Print found titles. */
	if flags.Debug {
		maxTitleWidth := len(ui.GetLongestTitle(titles))

		fmt.Println("\n\nFOUND TITLES:")
		for _, t := range titles {
			fmt.Printf("%-*s -> %s\n", maxTitleWidth, t.Name, t.Url)
		}
		fmt.Printf("\n\n")
	}

	return titles
}

/* Given a URL `searchURL`, return all titles found by FanCaps. */
func scrapeTitles(searchURL string, flags cli.CLIFlags) []*types.Title {
	var titles []*types.Title

	scraperOpts := GetScraperOpts(flags)
	c := colly.NewCollector(scraperOpts...)

	/* Extract title info. */
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Attr("href"))
		category := getCategory(url)
		title := &types.Title{
			Category: category,
			Name:     e.Text,
			Url:      url,
			Images:   &types.Images{},
		}
		titles = append(titles, title)
	})

	if flags.Debug {
		c.OnRequest(func(req *colly.Request) {
			fmt.Printf("SEARCH QUERY URL: %s\n", req.URL.String())
		})
	}

	c.Visit(searchURL)

	if !flags.NoAsync {
		c.Wait()
	}

	return titles
}

/* Return the category of a title based on its URL, `url`. */
func getCategory(url string) types.Category {
	switch {
	case strings.Contains(url, "/movies/"):
		return types.CategoryMovie
	case strings.Contains(url, "/tv/"):
		return types.CategoryTV
	case strings.Contains(url, "/anime/"):
		return types.CategoryAnime
	default:
		fmt.Fprintf(os.Stderr, "getCategory: couldn't extract category from url %s", url)
		os.Exit(1)
		return -1
	}
}

/*
Returns the season number from the name `name`
and whether a season number was found.
*/
func getSeasonNumber(name string) (int, bool) {
	match := seasonRegex.FindStringSubmatch(name)
	if match != nil {
		if n, err := strconv.Atoi(match[1]); err == nil {
			return n, true
		}
	}

	return 0, false
}
