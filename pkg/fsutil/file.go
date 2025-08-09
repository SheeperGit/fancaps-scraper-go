package fsutil

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
)

/*
Returns whether the image at URL `url` exists in the directory `imgDir`,
as well as the full image path that was checked.
*/
func ImageExists(imgDir string, url string) (bool, string) {
	imgFilename := path.Base(url)
	imgPath := filepath.Join(imgDir, imgFilename)

	if _, err := os.Stat(imgPath); err == nil {
		return true, imgPath
	}

	return false, imgPath
}

/* Returns a safe filename for file creation. */
func sanitizeFilename(filename string) string {
	/*
		Windows forbidden chars: \ / : * ? " < > |
		Get rid of spaces too.
	*/
	var forbiddenChars = regexp.MustCompile(`[\\/:*?"<>| ]`)

	return forbiddenChars.ReplaceAllString(filename, "_")
}
