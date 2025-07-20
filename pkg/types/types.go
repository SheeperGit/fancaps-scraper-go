package types

import (
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
