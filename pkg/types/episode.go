package types

import "time"

/* An episode of a title. */
type Episode struct {
	Title  *Title    // Title to which the episode belongs to.
	Name   string    // Name of the episode.
	Link   string    // Link to the episode on fancaps.net.
	Images *Images   // Image info about the episode. (Non-empty for Anime/TV Series Only)
	Start  time.Time // Start time of episode download.
}

/*
Increments the downloaded image counter of episode `e` and its title by 1,
as well as the global downloaded image counter across all titles.
*/
func (e *Episode) IncrementDownloaded() {
	e.Images.mu.Lock()
	defer e.Images.mu.Unlock()

	e.Images.downloaded++
	e.Title.IncrementDownloaded()
}

/*
Increments the skipped image counter of episode `e` and its title by 1,
as well as the global skipped image counter across all titles.
*/
func (e *Episode) IncrementSkipped() {
	e.Images.mu.Lock()
	defer e.Images.mu.Unlock()

	e.Images.skipped++
	e.Title.IncrementSkipped()
}

/*
Increments the total image counter of episode `e` and its title by 1,
as well as the global total image counter across all titles.
*/
func (e *Episode) IncrementImageTotal() {
	e.Images.mu.Lock()
	defer e.Images.mu.Unlock()

	e.Images.total++
	e.Title.IncrementImageTotal()
}
