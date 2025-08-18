package types

import "time"

/* An episode of a title. */
type Episode struct {
	Title  *Title    // Title to which the episode belongs to.
	Name   string    // Name of the episode.
	Url    string    // URL to the episode on fancaps.net.
	Images *Images   // Image info about the episode. (Non-empty for Anime/TV Series Only)
	Start  time.Time // Start time of episode download.
}

/* Returns the name of the episode `e`. */
func (e *Episode) GetName() string {
	return e.Name
}

/* Returns the title to which the episode `e` belongs to. */
func (e *Episode) GetTitle() *Title {
	return e.Title
}

/* Returns the start time of the download of episode `e`. */
func (e *Episode) GetStart() time.Time {
	return e.Start
}

/* Returns whether all the downloads of images of episode `e` are done. */
func (e *Episode) GetDone() bool {
	return e.Images.Done
}

/* Returns the number of downloaded images for episode `e`. */
func (e *Episode) Downloaded() uint32 {
	return e.Images.Downloaded()
}

/* Returns the number of skipped images for episode `e`. */
func (e *Episode) Skipped() uint32 {
	return e.Images.Skipped()
}

/* Returns the total number of images for episode `e`. */
func (e *Episode) Total() uint32 {
	return e.Images.Total()
}

/* Marks the download of episode `e` as done.  */
func (e *Episode) MarkDone() {
	e.Images.Done = true
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
