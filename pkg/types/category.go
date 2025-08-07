package types

/* Enum for Categories. */
type Category int

const (
	CategoryAnime Category = iota // Anime category.
	CategoryTV                    // TV Series category.
	CategoryMovie                 // Movie category.
)

var CategoryName = map[Category]string{
	CategoryAnime: "Anime",
	CategoryTV:    "TV Series",
	CategoryMovie: "Movies",
}

/* Convert a category enumeration to its corresponding string representation. */
func (cat Category) String() string {
	return CategoryName[cat]
}

/* Category statistics. */
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
