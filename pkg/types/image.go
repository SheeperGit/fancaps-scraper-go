package types

import (
	"sync"
	"time"
)

/* Image info on either a title or episode. */
type Images struct {
	urls       []string     // List of URLs to the images of a title or one of its episodes.
	downloaded uint32       // Amount of images downloaded.
	skipped    uint32       // Amount of images skipped.
	total      uint32       // Amount of images associated with a title or episode.
	Done       bool         // If true, all images are processed.
	mu         sync.RWMutex // Prevents bad writes from concurrent increments, while allowing multiple readers.
}

type ImageContainer interface {
	GetName() string
	GetTitle() *Title
	GetStart() time.Time
	GetDone() bool
	Downloaded() uint32
	Skipped() uint32
	Total() uint32
	MarkDone()
	IncrementDownloaded()
	IncrementSkipped()
	IncrementImageTotal()
}

/* Returns the URLs of the images `imgs`. */
func (imgs *Images) URLs() []string {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.urls
}

/* Returns the number of downloaded images for `imgs`. */
func (imgs *Images) Downloaded() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.downloaded
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
