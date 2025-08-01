package types

import (
	"sync"
	"time"
)

var (
	imgsProcessed uint32 // Total number of processed images.
	imgsSkipped   uint32 // Total number of skipped images.
	imgsTotal     uint32 // Total number of images.
)

var (
	skippedMu   sync.RWMutex // Prevents bad writes to `imgsSkipped`, while allowing multiple readers.
	processedMu sync.RWMutex // Prevents bad writes to `imgsProcessed`, while allowing multiple readers.
	totalMu     sync.RWMutex // Prevents bad writes to `imgsTotal`, while allowing multiple readers.
)

/* A Movie, TV Series, or Anime title. */
type Title struct {
	Episodes []*Episode // Episodes of the title.
	Category Category   // Category of the title.
	Name     string     // Name of the title.
	Link     string     // Link to the title on fancaps.net.
	Images   *Images    // Image info about the title. (Non-empty for Movie Titles Only)
	Start    time.Time  // Start time of title download.
}

/* An episode of a title. */
type Episode struct {
	Name   string    // Name of the episode.
	Link   string    // Link to the episode on fancaps.net.
	Images *Images   // Image info about the episode. (Non-empty for Anime/TV Series Only)
	Start  time.Time // Start time of episode download.
}

/* Image info on either a title or episode. */
type Images struct {
	urls      []string     // List of URLs to the images of a title or one of its episodes.
	processed uint32       // Amount of images processed. (Downloaded, skipped, or errored out.)
	skipped   uint32       // Amount of images skipped.
	total     uint32       // Amount of images associated with a title or episode.
	Done      bool         // If true, all images are processed.
	mu        sync.RWMutex // Prevents bad writes from concurrent increments, while allowing multiple readers.
}

/* Enum for Categories. */
type Category int

const (
	CategoryAnime Category = iota
	CategoryTV
	CategoryMovie
)

/* Convert a category enumeration to its corresponding string representation. */
func (cat Category) String() string {
	return CategoryName[cat]
}

var CategoryName = map[Category]string{
	CategoryAnime: "Anime",
	CategoryTV:    "TV Series",
	CategoryMovie: "Movies",
}

/* Thread-safe category amounts. */
type CatStats struct {
	Amts map[Category]int // Amount of titles per category.
	Max  int              // Highest amount of titles from all categories.
}

/* Returns category statistics of titles `titles`. */
func GetCatStats(titles []*Title) *CatStats {
	cs := &CatStats{
		Amts: make(map[Category]int, len(CategoryName)),
	}

	/* Count up titles per category while also keeping track of the maximum. */
	for _, title := range titles {
		cat := title.Category
		cs.Amts[cat]++
		if cs.Amts[cat] > cs.Max {
			cs.Max = cs.Amts[cat]
		}
	}

	return cs
}

/* Returns a list of categories with at least one found title. */
func (cs *CatStats) UsedCategories() []Category {
	var usedCats []Category
	for cat := Category(0); cat < Category(len(CategoryName)); cat++ {
		if cs.Amts[cat] != 0 {
			usedCats = append(usedCats, cat)
		}
	}

	return usedCats
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

/* Returns the total number of processed images across all titles. */
func ProcessedTotal() uint32 {
	processedMu.RLock()
	defer processedMu.RUnlock()

	return imgsProcessed
}

/* Returns the number of skipped images. */
func (imgs *Images) Skipped() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.skipped
}

/* Returns the total number of skipped images. */
func SkippedTotal() uint32 {
	skippedMu.RLock()
	defer skippedMu.RUnlock()

	return imgsSkipped
}

/* Returns the number of images. */
func (imgs *Images) Total() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.total
}

/* Returns the total number of images. */
func TotalTotal() uint32 {
	totalMu.RLock()
	defer totalMu.RUnlock()

	return imgsTotal
}
