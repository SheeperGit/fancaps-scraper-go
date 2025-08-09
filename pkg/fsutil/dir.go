package fsutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

/*
Returns the path to a newly created output directory at `dirname` to store the scraped images.
This function checks whether the parent directories of `dirname` exist before creating the directory,
if they do not, this will exit with code 1.

Anime images will be saved to "./`dirname`/<Anime_Title_Name>/<Anime_Episode_Name>/".

TV Series images will be saved to "./`dirname`/<TV_Title_Name>/<TV_Episode_Name>/".

Movie images will be saved to "./`dirname`/<Movie_Name>/".
*/
func CreateOutputDir(dirname string) string {
	/* Check (for a second time) that the parent directories still exist. */
	if !ParentDirsExist(dirname) {
		fmt.Fprintf(os.Stderr,
			ui.ErrStyle.Render("CreateOutputDir error: Couldn't find parent directories of `%s`")+"\n"+
				ui.ErrStyle.Render("Make sure the parent directories still exist at runtime.")+"\n",
			dirname)
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
func CreateTitleDir(outDir string, titleName string) string {
	sanitizedTitleName := sanitizeFilename(titleName)

	titleDir := filepath.Join(outDir, sanitizedTitleName)
	mkdirIfDNE(titleDir)

	return titleDir
}

/*
Creates a new directory for an episode under the name `episodeName` in the directory `titleDir`.
The `titleDir` directory must exist.
Returns the path to the newly created episode directory.
*/
func CreateEpisodeDir(titleDir string, episodeName string) string {
	sanitizedEpisodeName := sanitizeFilename(episodeName)

	episodeDir := filepath.Join(titleDir, sanitizedEpisodeName)
	mkdirIfDNE(episodeDir)

	return episodeDir
}

/*
Returns true, if the parent directories of `dirPath` exist
and returns false otherwise.
*/
func ParentDirsExist(dirPath string) bool {
	parentDirs := filepath.Dir(dirPath)

	info, err := os.Stat(parentDirs)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Fprintf(os.Stderr, "ParentDirsExist unexpected error: %v", err)
		return false
	}

	return info.IsDir()
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
