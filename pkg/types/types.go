package types

import (
	"maps"
	"sync"
)

/* A Movie, TV Series, or Anime title. */
type Title struct {
	Episodes []Episode // Episodes of the title.
	Category Category  // Category of the title.
	Name     string    // Name of the title.
	Link     string    // Link to the title on fancaps.net.
}

/* An episode of a title. */
type Episode struct {
	Images []string // URLs pointing to the images of the screencaps.
	Name   string   // Name of the episode.
	Link   string   // Link to the episode on fancaps.net.
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
	mu   sync.Mutex       // Prevents bad writes from concurrent increments.
	amts map[Category]int // Amount of titles per category.
}

/* Returns a new category statistics struct. */
func NewCatStats() *CatStats {
	return &CatStats{
		amts: make(map[Category]int, len(CategoryName)),
	}
}

/* Increments category `cat` by 1. */
func (m *CatStats) Increment(cat Category) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.amts[cat]++
}

/* Returns the amount of titles found for category `cat`. */
func (m *CatStats) Get(cat Category) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.amts[cat]
}

/* Returns a copy of the category amounts. */
func (m *CatStats) Snapshot() map[Category]int {
	m.mu.Lock()
	defer m.mu.Unlock()

	copy := make(map[Category]int, len(m.amts))
	maps.Copy(copy, m.amts)

	return copy
}

/* Returns the highest amount of titles from all categories. */
func (cs *CatStats) Max() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	max := 0
	for _, v := range cs.amts {
		if v > max {
			max = v
		}
	}

	return max
}

/* Returns a list of categories with at least one found title. */
func (c *CatStats) UsedCategories() []Category {
	c.mu.Lock()
	defer c.mu.Unlock()

	var usedCats []Category
	for cat := Category(0); cat < Category(len(CategoryName)); cat++ {
		if c.amts[cat] != 0 {
			usedCats = append(usedCats, cat)
		}
	}

	return usedCats
}
