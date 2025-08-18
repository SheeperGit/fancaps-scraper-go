package format

import (
	"encoding/json"

	"sheeper.com/fancaps-scraper-go/pkg/types"
)

type JSONOutput struct {
	Total  int         `json:"total"`
	Titles []JSONTitle `json:"titles"`
}

/* A Title JSON object. */
type JSONTitle struct {
	Name     string        `json:"name"`
	Category string        `json:"category"`
	Url      string        `json:"url"`
	Episodes []JSONEpisode `json:"episodes"`
	Images   []string      `json:"images,omitempty"`
}

/* An Episode JSON object. */
type JSONEpisode struct {
	Name   string   `json:"name"`
	Url    string   `json:"url"`
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
			Url:      t.Url,
			Images:   t.Images.URLs(),
		}
		for _, ep := range t.Episodes {
			jsonTitle.Episodes = append(jsonTitle.Episodes, JSONEpisode{
				Name:   ep.Name,
				Url:    ep.Url,
				Images: ep.Images.URLs(),
			})
		}
		jsonTitles = append(jsonTitles, jsonTitle)
	}

	output := JSONOutput{
		Total:  len(jsonTitles),
		Titles: jsonTitles,
	}

	return json.MarshalIndent(output, "", "  ")
}

/* Returns the content type of the JSON formatter. */
func (JSONFormatter) ContentType() string {
	return "application/json"
}
