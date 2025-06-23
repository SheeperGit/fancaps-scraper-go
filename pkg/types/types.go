package types

/* A Movie, TV Series, or Anime title. */
type Title struct {
	Episodes []Episode
	Category Category
	Name     string
	Link     string
}

/* An episode of a title. */
type Episode struct {
	Images []string
	Name   string
	Link   string
}

/* Enum for Categories. */
type Category int

const (
	CategoryMovie Category = iota
	CategoryTV
	CategoryAnime
	CategoryUnknown
)

var CategoryName = map[Category]string{
	CategoryMovie:   "Movies",
	CategoryTV:      "TV Series",
	CategoryAnime:   "Anime",
	CategoryUnknown: "Category Unknown",
}

/* Convert a category enumeration to its corresponding string representation. */
func (cat Category) String() string {
	return CategoryName[cat]
}
