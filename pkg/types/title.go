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

/* Returns the name of the title `t`. */
func (t *Title) GetName() string {
	return t.Name
}

/* Returns the title to which the title `t` belongs to. (i.e., returns itself) */
func (t *Title) GetTitle() *Title {
	return t
}

/* Returns the start time of the download of title `t`. */
func (t *Title) GetStart() time.Time {
	return t.Start
}

/* Returns whether all the downloads of images of title `t` are done. */
func (t *Title) GetDone() bool {
	return t.Images.Done
}

/* Returns the number of downloaded images from title `t`. */
func (t *Title) Downloaded() uint32 {
	var downloaded uint32
	for _, e := range t.Episodes {
		downloaded += e.Downloaded()
	}

	return downloaded
}

/* Returns the number of skipped images from title `t`. */
func (t *Title) Skipped() uint32 {
	var skipped uint32
	for _, e := range t.Episodes {
		skipped += e.Skipped()
	}

	return skipped
}

/* Returns the total number of images from title `t`. */
func (t *Title) Total() uint32 {
	var total uint32
	for _, e := range t.Episodes {
		total += e.Total()
	}

	return total
}

/* Marks the download of title `t` as done.  */
func (t *Title) MarkDone() {
	t.Images.Done = true
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
