package scraper

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sheeper.com/fancaps-scraper-go/pkg/cli"
	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui/progressbar"
)

const (
	defaultMaxWorkers = 3    // Default maximum amount of titles or episodes to download images from in parallel.
	defaultMinDelay   = 1000 // Default minimum delay (in milliseconds) after every new image download request.
	defaultRandDelay  = 5000 // Default maximum random delay (in milliseconds) after every new image download request.
)

var (
	launchedWorkers = 0        // Number of image download workers launched.
	launchMu        sync.Mutex // Prevents bad writes to `launchedWorkers`.
)

/* Download images from titles `titles`. */
func DownloadImages(titles []*types.Title, flags cli.CLIFlags) {
	sema := make(chan struct{}, defaultMaxWorkers)
	var wg sync.WaitGroup
	outputDir := createOutputDir(flags.OutputDir)

	downloadImg := func(imgDir string, url string, titleImages, episodeImages *types.Images) {
		sent := downloadImage(imgDir, url)

		if episodeImages != nil {
			episodeImages.IncrementAmtProcessed()
		}
		titleImages.IncrementAmtProcessed()

		progressbar.ShowProgress(titles)

		/* Only delay the next image request if one was sent in the first place. */
		if sent {
			jitterDelay(defaultMinDelay, defaultRandDelay)
		}
	}

	downloadImgAsync := func(imgDir, url string, titleImages, episodeImages *types.Images) {
		wg.Add(1)
		sema <- struct{}{}
		go func(url string) {
			/* Update the number of launched workers. */
			launchMu.Lock()
			workerNum := launchedWorkers
			if workerNum < defaultMaxWorkers {
				launchedWorkers++
			}
			launchMu.Unlock()

			/* Delay the workers slightly for the first time. */
			if workerNum < defaultMaxWorkers {
				jitterDelay(workerNum*500, 1000)
			}

			defer wg.Done()
			defer func() { <-sema }()
			downloadImg(imgDir, url, titleImages, episodeImages)
		}(url)
	}

	fmt.Println(":: Showing progress...")

	/* For each title... */
	for _, title := range titles {
		titleDir := createTitleDir(outputDir, title.Name)

		/* Handle movies seperately, since they have no episodes. */
		if title.Category == types.CategoryMovie {
			URLs := title.Images.GetImages()
			for _, url := range URLs {
				if flags.Async {
					downloadImgAsync(titleDir, url, title.Images, nil)
				} else {
					downloadImg(titleDir, url, title.Images, nil)
				}
			}

			continue // Go to next title.
		}

		/* For each episode... */
		for _, episode := range title.Episodes {
			imgDir := createEpisodeDir(titleDir, episode.Name)

			URLs := episode.Images.GetImages()
			for _, url := range URLs {
				if flags.Async {
					downloadImgAsync(imgDir, url, title.Images, episode.Images)
				} else {
					downloadImg(imgDir, url, title.Images, episode.Images)
				}
			}
		}
	}

	if flags.Async {
		wg.Wait()
	}
}

/*
Sleeps for a minimum of `minDelay` milliseconds and a random amount
of milliseconds ranging from 0 milliseconds (no random delay) to
`randDelay` milliseconds. Returns the amount of time slept.

In this way, `randDelay` acts as the maximum amount of random delay possible
(in milliseconds).
*/
func jitterDelay(minDelay int, randDelay int) time.Duration {
	d := time.Duration(minDelay) * time.Millisecond
	r := time.Duration(rand.Intn(randDelay)) * time.Millisecond
	jitter := d + r

	time.Sleep(jitter)

	return jitter
}

/*
Returns the path to a newly created output directory at `dirname` to store the scraped images.
This function checks whether the parent directories of `dirname` exist before creating the directory,
if they do not, this will exit with code 1.

Anime images will be saved to "./`dirname`/<Anime_Title_Name>/<Anime_Episode_Name>/".

TV Series images will be saved to "./`dirname`/<TV_Title_Name>/<TV_Episode_Name>/".

Movie images will be saved to "./`dirname`/<Movie_Name>/".
*/
func createOutputDir(dirname string) string {
	/* Check (for a second time) that the parent directories still exist. */
	if !cli.ParentDirsExist(dirname) {
		fmt.Fprintf(os.Stderr, "createOutputDir error: Couldn't find parent directories of '%s'\n", dirname)
		fmt.Fprintf(os.Stderr, "Make sure the parent directories still exist at runtime.\n")
		os.Exit(1)
	}

	mkdirIfDNE(dirname)

	return dirname
}

/*
Creates a new directory for a title under the name `titleName` in the directory `outDir`.
The `outDir` directory must exist.
Returns the path to the newly created title directory.
*/
func createTitleDir(outDir string, titleName string) string {
	sanitizedTitleName := sanitizeDirname(titleName)

	titleDir := filepath.Join(outDir, sanitizedTitleName)
	mkdirIfDNE(titleDir)

	return titleDir
}

/*
Creates a new directory for an episode under the name `episodeName` in the directory `titleDir`.
The `titleDir` directory must exist.
Returns the path to the newly created episode directory.
*/
func createEpisodeDir(titleDir string, episodeName string) string {
	sanitizedEpisodeName := sanitizeDirname(episodeName)

	episodeDir := filepath.Join(titleDir, sanitizedEpisodeName)
	mkdirIfDNE(episodeDir)

	return episodeDir
}

/*
Downloads the image found at `url` to the directory `imgDir`,
and returns whether the request to `url` was made.

Although not strictly enforced, `imgDir` is expected to refer to an "Episode directory"
for Anime and TV Series titles or a "Title directory" for Movie titles.
Prints errors for locating the image, file creation, or copying content to a file, if encountered.
*/
func downloadImage(imgDir string, url string) bool {
	imgFilename := path.Base(url)
	imgPath := filepath.Join(imgDir, imgFilename)
	sent := false

	/* If file already exists, don't overwrite and print an error. */
	if _, err := os.Stat(imgPath); err == nil {
		// fmt.Fprintf(os.Stderr, "Skipping existing file: %s\n", imgPath)
		return sent
	} else if !os.IsNotExist(err) {
		// fmt.Fprintf(os.Stderr, "Failed to stat file (%s): %v\n", imgPath, err)
		return sent
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Failed to create HTTP request: %v\n", err)
		return sent
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
	req.Header.Set("Referer", "https://fancaps.net")

	client := &http.Client{}
	res, err := client.Do(req)
	sent = true
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Failed to perform HTTP request: %v\n", err)
		return sent
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		fmt.Fprintln(os.Stderr, "You are being rate-limited. Try again later.")
		fmt.Fprintln(os.Stderr, "Hint: Try setting `--max-download-threads` to a lower value.")
		os.Exit(2)
	} else if res.StatusCode != http.StatusOK {
		// fmt.Fprintf(os.Stderr, "Bad status code: %d for URL: %s\n", res.StatusCode, url)
		return sent
	}

	/* Open file to copy image contents to. */
	file, err := os.Create(imgPath)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Failed to create file (%s): %v\n", imgPath, err)
		return sent
	}
	defer file.Close()

	/* Copy the response body to the file. */
	_, err = io.Copy(file, res.Body)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Failed to copy image contents to file (%s): %v\n", imgPath, err)
		return sent
	}

	return sent
}

/* Returns a safe directory name for directory creation on all platforms. */
func sanitizeDirname(dirname string) string {
	sanitizedDirname := strings.ReplaceAll(dirname, " ", "_")          // underscores are nicer than whitespaces :)
	sanitizedDirname = strings.ReplaceAll(sanitizedDirname, "/", "-")  // avoid nested paths
	sanitizedDirname = strings.ReplaceAll(sanitizedDirname, ":", "__") // remove illegal characters (Windows)

	return sanitizedDirname
}

/*
Creates directory `dirname`, if it does not already exist.
If directory creation fails, prints an error and exits with code 1.
*/
func mkdirIfDNE(dirname string) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if err := os.Mkdir(dirname, os.ModePerm); err != nil {
			log.Fatalf("mkdirIfDNE error: %v", err)
		}
	}
}
