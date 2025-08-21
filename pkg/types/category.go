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
