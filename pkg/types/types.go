package types

import (
	"maps"
	"sync"
)

/* A Movie, TV Series, or Anime title. */
type Title struct {
	Episodes []*Episode // Episodes of the title.
	Category Category   // Category of the title.
	Name     string     // Name of the title.
	Link     string     // Link to the title on fancaps.net.
	Images   *Images    // Image info about the title. (Non-empty for Movie Titles Only)
}

/* An episode of a title. */
type Episode struct {
	Name   string  // Name of the episode.
	Link   string  // Link to the episode on fancaps.net.
	Images *Images // Image info about the episode. (Non-empty for Anime/TV Series Only)
}

/* Image info on either a title or episode. */
type Images struct {
	URLs         []string     // List of URLs to the images of a title or one of its episodes.
	ImgCount     uint32       // Amount of images associated with a title or episode.
	AmtProcessed uint32       // Amount of images processed. (Downloaded, skipped, or errored out.)
	mu           sync.RWMutex // Prevents bad writes from concurrent increments, while allowing multiple readers.
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
	amts map[Category]int // Amount of titles per category.
	mu   sync.RWMutex     // Prevents bad writes from concurrent increments, while allowing multiple readers.
}

/* Returns a new category statistics struct. */
func NewCatStats() *CatStats {
	return &CatStats{
		amts: make(map[Category]int, len(CategoryName)),
	}
}

/* Increments category `cat` by 1. */
func (cs *CatStats) Increment(cat Category) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.amts[cat]++
}

/* Returns the amount of titles found for category `cat`. */
func (cs *CatStats) Get(cat Category) int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return cs.amts[cat]
}

/* Returns a copy of the category amounts. */
func (cs *CatStats) Snapshot() map[Category]int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	copy := make(map[Category]int, len(cs.amts))
	maps.Copy(copy, cs.amts)

	return copy
}

/* Returns the highest amount of titles from all categories. */
func (cs *CatStats) Max() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	max := 0
	for _, v := range cs.amts {
		if v > max {
			max = v
		}
	}

	return max
}

/* Returns a list of categories with at least one found title. */
func (cs *CatStats) UsedCategories() []Category {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var usedCats []Category
	for cat := Category(0); cat < Category(len(CategoryName)); cat++ {
		if cs.amts[cat] != 0 {
			usedCats = append(usedCats, cat)
		}
	}

	return usedCats
}

/* Increments image count by 1. */
func (imgs *Images) IncrementImgCount() {
	imgs.mu.Lock()
	defer imgs.mu.Unlock()

	imgs.ImgCount++
}

/* Increments image count by 1. */
func (imgs *Images) IncrementAmtProcessed() {
	imgs.mu.Lock()
	defer imgs.mu.Unlock()

	imgs.AmtProcessed++
}

/* Adds a URL. */
func (imgs *Images) AddURL(url string) {
	imgs.mu.Lock()
	defer imgs.mu.Unlock()

	imgs.URLs = append(imgs.URLs, url)
}

/* Returns the URLs of images found. */
func (imgs *Images) GetImages() []string {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.URLs
}

/* Returns the amount of images found. */
func (imgs *Images) GetImgCount() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.ImgCount
}

/* Returns the URLs of images found. */
func (imgs *Images) GetAmtProcessed() uint32 {
	imgs.mu.RLock()
	defer imgs.mu.RUnlock()

	return imgs.AmtProcessed
}
