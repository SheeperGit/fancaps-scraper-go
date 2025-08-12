package format

import (
	"encoding/json"

	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* A Title JSON object. */
type JSONTitle struct {
	Name     string        `json:"name"`
	Category string        `json:"category"`
	Link     string        `json:"link"`
	Episodes []JSONEpisode `json:"episodes"`
	Images   []string      `json:"images,omitempty"`
}

/* An Episode JSON object. */
type JSONEpisode struct {
	Name   string   `json:"name"`
	Link   string   `json:"link"`
	Images []string `json:"images,omitempty"`
}

type JSONFormatter struct{}

var jsonFmt = JSONFormatter{}

/* Returns a JSON representation of titles `titles`. */
func (JSONFormatter) Format(titles []*types.Title) ([]byte, error) {
	var jsonTitles []JSONTitle
	for _, t := range titles {
		jsonTitle := JSONTitle{
			Name:     t.Name,
			Category: t.Category.String(),
			Link:     t.Link,
			Images:   t.Images.URLs(),
		}
		for _, ep := range t.Episodes {
			jsonTitle.Episodes = append(jsonTitle.Episodes, JSONEpisode{
				Name:   ep.Name,
				Link:   ep.Link,
				Images: ep.Images.URLs(),
			})
		}
		jsonTitles = append(jsonTitles, jsonTitle)
	}

	return json.MarshalIndent(jsonTitles, "", "  ")
}

/* Returns the content type of the JSON formatter. */
func (JSONFormatter) ContentType() string {
	return "application/json"
}
