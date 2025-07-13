package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const defaultOutputDir = "output"

/*
Returns the path to a newly created output directory at `dirname` to store the scraped images.

Anime images will be saved to "./dirname/<Anime_Title_Name>/<Anime_Episode_Name>/".

TV Series images will be saved to "./dirname/<TV_Title_Name>/<TV_Episode_Name>/".

Movie images will be saved to "./dirname/<Movie_Name>/".
*/
func createOutputDir(dirname string) string {
	// TODO: Allow user to specify path and check that it exists on flag creation.
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		dirname = defaultOutputDir
	}

	outputPath := filepath.Join(".", dirname)
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return outputPath
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
Saves the image found at `url` to the directory `imgDir`.
Although not strictly enforced, `imgDir` is expected to refer to an "Episode directory"
for Anime and TV Series titles or a "Title directory" for Movie titles.
Prints errors for locating the image, file creation, or copying content to a file, if encountered.
*/
func saveImage(imgDir string, url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create HTTP request: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
	req.Header.Set("Referer", "https://fancaps.net")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get HTTP response: %v\n", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Bad status code: %d for URL: %s\n", res.StatusCode, url)
		return
	}

	imgFilename := path.Base(url)
	imgPath := filepath.Join(imgDir, imgFilename)

	/* If file already exists, don't overwrite and print an error. */
	if _, err := os.Stat(imgPath); err == nil {
		fmt.Fprintf(os.Stderr, "Skipping existing file: %s\n", imgPath)
		return
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Failed to stat file (%s): %v\n", imgPath, err)
		return
	}

	/* Open file to copy image contents to. */
	file, err := os.Create(imgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file (%s): %v\n", imgPath, err)
		return
	}
	defer file.Close()

	/* Copy the response body to the file. */
	_, err = io.Copy(file, res.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to copy image contents to file (%s): %v\n", imgPath, err)
		return
	}

	fmt.Printf("%s downloaded successfully!\n", imgPath)
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
If directory creation fails, prints an error and exits.
*/
func mkdirIfDNE(dirname string) {
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if err := os.Mkdir(dirname, os.ModePerm); err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}
}
