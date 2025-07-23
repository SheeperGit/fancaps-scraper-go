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

var (
	logFilename string    // Filename of the log file. Contains errors of varying severity.
	setLogFile  sync.Once // Sets the name of the log file.
)

/* Enum for error severity. */
type errSeverity int

const (
	ERR_WARNING errSeverity = iota
	ERR_ERROR
)

/* Convert a category enumeration to its corresponding string representation. */
func (es errSeverity) String() string {
	return severityName[es]
}

var severityName = map[errSeverity]string{
	ERR_WARNING: "WARNING", // Non-critical error severity.
	ERR_ERROR:   "ERROR",   // Critical error severity.
}

/* Maximum length string of a severity error. */
var maxSeverityLen = func() int {
	max := 0
	for _, name := range severityName {
		if len(name) > max {
			max = len(name)
		}
	}
	return max
}()

/* Download images from titles `titles`. */
func DownloadImages(titles []*types.Title, flags cli.CLIFlags) {
	var wg sync.WaitGroup
	sema := make(chan struct{}, flags.ParallelDownloads)

	downloadImg := func(imgDir string, url string, titleImages, episodeImages *types.Images) {
		if imageExists(imgDir, url) {
			progressbar.UpdateProgressDisplay(titles, titleImages, episodeImages)
			return
		}

		/* Initial delay. */
		jitterDelay(flags.MinDelay/2, flags.RandDelay/2)

		sent := downloadImage(imgDir, url)

		progressbar.UpdateProgressDisplay(titles, titleImages, episodeImages)

		/* Post delay. Only delay the next image request, if one was sent in the first place. */
		if sent {
			jitterDelay(flags.MinDelay/2, flags.RandDelay/2)
		}
	}

	downloadImgAsync := func(imgDir, url string, titleImages, episodeImages *types.Images) {
		wg.Add(1)
		sema <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() { <-sema }()

			downloadImg(imgDir, url, titleImages, episodeImages)
		}(url)
	}

	outputDir := createOutputDir(flags.OutputDir)

	fmt.Println(":: Showing progress...")
	progressbar.ShowProgress(titles)

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
func jitterDelay(minDelay uint32, randDelay uint32) time.Duration {
	d := time.Duration(minDelay) * time.Millisecond
	r := time.Duration(0)
	if randDelay > 0 {
		r = time.Duration(rand.Intn(int(randDelay))) * time.Millisecond
	}
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
		logErrorf(ERR_WARNING, "Skipping existing file: %s", imgPath)
		return sent
	} else if !os.IsNotExist(err) {
		logErrorf(ERR_WARNING, "Failed to stat file (%s): %v", imgPath, err)
		return sent
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logErrorf(ERR_ERROR, "Failed to create HTTP request: %v", err)
		return sent
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
	req.Header.Set("Referer", "https://fancaps.net")

	client := &http.Client{}
	res, err := client.Do(req)
	sent = true
	if err != nil {
		logErrorf(ERR_ERROR, "Failed to perform HTTP request: %v", err)
		return sent
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusForbidden {
		fmt.Fprintln(os.Stderr, "You are being rate-limited. Try again later.")
		fmt.Fprintln(os.Stderr, "Hint: Try setting `--parallel-downloads` to a lower value.")
		os.Exit(2)
	} else if res.StatusCode != http.StatusOK {
		logErrorf(ERR_ERROR, "Bad status code: %d for URL: %s", res.StatusCode, url)
		return sent
	}

	/* Open file to copy image contents to. */
	file, err := os.Create(imgPath)
	if err != nil {
		logErrorf(ERR_ERROR, "Failed to create file (%s): %v", imgPath, err)
		return sent
	}
	defer file.Close()

	/* Copy the response body to the file. */
	_, err = io.Copy(file, res.Body)
	if err != nil {
		logErrorf(ERR_ERROR, "Failed to copy image contents to file (%s): %v", imgPath, err)
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

/*
Appends errors to a log file, as defined by its severity `severity`, format `format` and its
arguments `args`. Errors are timestamped with nanosecond precision.
*/
func logErrorf(severity errSeverity, format string, args ...any) {
	setLogFile.Do(func() {
		fileTimestamp := time.Now().Format("2006-01-02_15-04-05.000000000") // Nanosecond precision.
		logFilename = fmt.Sprintf("fsg_errors_%s.txt", fileTimestamp)
	})

	f, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open error log: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	errTimestamp := time.Now().Format("2006-01-02 15:04:05.000000000")          // Nanosecond precision.
	sev := fmt.Sprintf("%-*s", maxSeverityLen+2, fmt.Sprintf("[%s]", severity)) // Left-align severity error text.
	errLine := fmt.Sprintf("%s (%s) %s\n", sev, errTimestamp, fmt.Sprintf(format, args...))
	f.WriteString(errLine)
}

/*
Returns true, if the image found at URL `url` exists in the directory `imgDir`
and returns false otherwise.
*/
func imageExists(imgDir string, url string) bool {
	imgFilename := path.Base(url)
	imgPath := filepath.Join(imgDir, imgFilename)

	if _, err := os.Stat(imgPath); err == nil {
		return true
	}

	return false
}
