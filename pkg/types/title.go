package types

import "time"

/* A Movie, TV Series, or Anime title. */
type Title struct {
	Episodes []*Episode // Episodes of the title.
	Category Category   // Category of the title.
	Name     string     // Name of the title.
	Link     string     // Link to the title on fancaps.net.
	Images   *Images    // Image info about the title. (Non-empty for Movie Titles Only)
	Start    time.Time  // Start time of title download.
}

/*
Increments the downloaded image counter of title `t` by 1,
as well as the global downloaded image counter across all titles.
*/
func (t *Title) IncrementDownloaded() {
	t.Images.mu.Lock()
	defer t.Images.mu.Unlock()

	t.Images.downloaded++
	IncrementGlobalDownloadedCount()
}

/*
Increments the skipped image counter of title `t` by 1,
as well as the global skipped image counter across all titles.
*/
func (t *Title) IncrementSkipped() {
	t.Images.mu.Lock()
	defer t.Images.mu.Unlock()

	t.Images.skipped++
	IncrementGlobalSkippedCount()
}

/*
Increments total image counter of title `t` by 1,
as well as the global total image counter across all titles.
*/
func (t *Title) IncrementImageTotal() {
	t.Images.mu.Lock()
	defer t.Images.mu.Unlock()

	t.Images.total++
	IncrementGlobalImageCount()
}
