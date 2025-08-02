package types

import "time"

/* An episode of a title. */
type Episode struct {
	Name   string    // Name of the episode.
	Link   string    // Link to the episode on fancaps.net.
	Images *Images   // Image info about the episode. (Non-empty for Anime/TV Series Only)
	Start  time.Time // Start time of episode download.
}
