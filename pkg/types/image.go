package types

import "sync"

/* Image info on either a title or episode. */
type Images struct {
	urls      []string     // List of URLs to the images of a title or one of its episodes.
	processed uint32       // Amount of images processed. (Downloaded, skipped, or errored out.)
	skipped   uint32       // Amount of images skipped.
	total     uint32       // Amount of images associated with a title or episode.
	Done      bool         // If true, all images are processed.
	mu        sync.RWMutex // Prevents bad writes from concurrent increments, while allowing multiple readers.
}

/* Returns the URLs of the images `imgs`. */
func (imgs *Images) URLs() []string {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.urls
}

/* Returns the number of processed images for `imgs`. */
func (imgs *Images) Processed() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.processed
}

/* Returns the number of skipped images. */
func (imgs *Images) Skipped() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.skipped
}

/* Returns the number of images. */
func (imgs *Images) Total() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.total
}

/* Adds a URL. */
func (imgs *Images) AddURL(url string) {
	imgs.mu.Lock()
	defer imgs.mu.Unlock()

	imgs.urls = append(imgs.urls, url)
}

/*
Increments the processed image counter of `imgs` by 1,
as well as the global processed image counter across all titles.
*/
func (imgs *Images) IncrementProcessed() {
	imgs.mu.Lock()
	processedMu.Lock()
	defer imgs.mu.Unlock()
	defer processedMu.Unlock()

	imgs.processed++ // Local processed counter.
	imgsProcessed++  // Global processed counter.
}

/*
Increments the skipped image counter of `imgs` by 1,
as well as the global skipped image counter across all titles.
*/
func (imgs *Images) IncrementSkipped() {
	imgs.mu.Lock()
	skippedMu.Lock()
	defer imgs.mu.Unlock()
	defer skippedMu.Unlock()

	imgs.skipped++ // Local skipped counter.
	imgsSkipped++  // Global skipped counter.
}

/* Increments total image counter of `imgs` by 1. */
func (imgs *Images) IncrementTotal() {
	imgs.mu.Lock()
	totalMu.Lock()
	defer imgs.mu.Unlock()
	defer totalMu.Unlock()

	imgs.total++ // Local total counter.
	imgsTotal++  // Global total counter.
}
