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
