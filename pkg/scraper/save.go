package scraper

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/fsutil"
	"sheeper.com/fancaps-scraper-go/pkg/logf"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui/progressbar"
)

/* Download images from titles `titles`. */
func DownloadImages(titles []*types.Title) {
	var wg sync.WaitGroup
	flags := cli.Flags()
	sema := make(chan struct{}, flags.ParallelDownloads)

	downloadImg := func(imgDir string, imgCon types.ImageContainer, url string) {
		if exists, imgPath := fsutil.ImageExists(imgDir, url); exists {
			logf.LogErrorf(logf.LOG_WARNING, "Skipping existing file: %s", imgPath)
			progressbar.UpdateProgressDisplay(titles, imgCon.IncrementSkipped)
			return
		}

		/* Pre-delay. */
		jitterDelay(flags.MinDelay/2, flags.RandDelay/2)

		sent := downloadImage(imgDir, url)

		progressbar.UpdateProgressDisplay(titles, imgCon.IncrementDownloaded)

		/* Post-delay. Only delay the next image request, if one was sent in the first place. */
		if sent {
			jitterDelay(flags.MinDelay/2, flags.RandDelay/2)
		}
	}

	downloadImgAsync := func(imgDir string, imgCon types.ImageContainer, url string) {
		wg.Add(1)
		sema <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() { <-sema }()

			downloadImg(imgDir, imgCon, url)
		}(url)
	}

	outputDir := fsutil.CreateOutputDir(flags.OutputDir)

	fmt.Println(":: Showing progress...")
	progressbar.ShowProgress(titles)

	/* For each title... */
	for _, title := range titles {
		titleDir := fsutil.CreateTitleDir(outputDir, title.Name)
		title.Start = time.Now()

		/* Handle movies seperately, since they have no episodes. */
		if title.Category == types.CategoryMovie {
			URLs := title.Images.URLs()
			for _, url := range URLs {
				if !flags.NoAsync {
					downloadImgAsync(titleDir, title, url)
				} else {
					downloadImg(titleDir, title, url)
				}
			}

			continue // Go to next title.
		}

		/* For each episode... */
		for _, episode := range title.Episodes {
			episodeDir := fsutil.CreateEpisodeDir(titleDir, episode.Name)

			URLs := episode.Images.URLs()
			episode.Start = time.Now()
			for _, url := range URLs {
				if !flags.NoAsync {
					downloadImgAsync(episodeDir, episode, url)
				} else {
					downloadImg(episodeDir, episode, url)
				}
			}
		}
	}

	if !flags.NoAsync {
		wg.Wait()
	}
}

/*
Downloads the image found at the URL `url` to the directory `imgDir`,
and returns whether the request to download the image was made.

Although not strictly enforced, `imgDir` is expected to refer to an "Episode directory"
for Anime and TV Series titles or a "Title directory" for Movie titles.
Logs errors for locating the image, file creation, or copying content to a file, if encountered.
*/
func downloadImage(imgDir string, url string) bool {
	imgFilename := path.Base(url)
	imgPath := filepath.Join(imgDir, imgFilename)
	sent := false

	/* If file already exists, don't overwrite and log as a error. */
	if _, err := os.Stat(imgPath); err == nil {
		logf.LogErrorf(logf.LOG_ERROR, "Inconsistent file state: %s was absent during initial check, but exists now", imgPath)
		return sent
	} else if !os.IsNotExist(err) {
		logf.LogErrorf(logf.LOG_ERROR, "Failed to stat file (%s): %v", imgPath, err)
		return sent
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logf.LogErrorf(logf.LOG_ERROR, "Failed to create HTTP request: %v", err)
		return sent
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
	req.Header.Set("Referer", "https://fancaps.net")

	client := &http.Client{}
	res, err := client.Do(req)
	sent = true
	if err != nil {
		logf.LogErrorf(logf.LOG_ERROR, "Failed to perform HTTP request: %v", err)
		return sent
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusForbidden {
		fmt.Fprintln(os.Stderr,
			"You are being rate-limited. Try again later."+"\n"+
				"Hint: Try setting `--parallel-downloads` to a lower value.")
		os.Exit(2)
	} else if res.StatusCode != http.StatusOK {
		logf.LogErrorf(logf.LOG_ERROR, "Bad status code: %d for URL: %s", res.StatusCode, url)
		return sent
	}

	/* Open file to copy image contents to. */
	file, err := os.Create(imgPath)
	if err != nil {
		logf.LogErrorf(logf.LOG_ERROR, "Failed to create file (%s): %v", imgPath, err)
		return sent
	}
	defer file.Close()

	/* Copy the response body to the file. */
	_, err = io.Copy(file, res.Body)
	if err != nil {
		logf.LogErrorf(logf.LOG_ERROR, "Failed to copy image contents to file (%s): %v", imgPath, err)
		return sent
	}

	return sent
}

/*
Sleeps for a minimum of `minDelay` time and a random amount
ranging from 0 (no random delay) to `randDelay` time.
Returns the amount of time slept.

In this way, `randDelay` acts as the maximum amount of random delay possible.
*/
func jitterDelay(minDelay, randDelay time.Duration) time.Duration {
	var r time.Duration
	if randDelay > 0 {
		r = time.Duration(rand.Int63n(int64(randDelay)))
	}
	jitter := minDelay + r

	time.Sleep(jitter)

	return jitter
}
