package format

import (
	"encoding/csv"
	"strings"

	"sheeper.com/fancaps-scraper-go/pkg/types"
)

type CSVFormatter struct{}

var csvFmt = CSVFormatter{}

var schema = []string{
	"Title Name",
	"Category",
	"Title Link",
	"Episode Name",
	"Episode Link",
	"Image URL",
}

/* Returns a CSV representation of titles `titles`. */
func (CSVFormatter) Format(titles []*types.Title) ([]byte, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)

	if err := w.Write(schema); err != nil {
		return nil, err
	}

	for _, t := range titles {
		if t.Category == types.CategoryMovie { // Handle movies seperately, since they have no episodes.
			for _, img := range t.Images.URLs() {
				row := []string{t.Name, t.Category.String(), t.Link, "", "", img}
				if err := w.Write(row); err != nil {
					return nil, err
				}
			}
		} else {
			for _, ep := range t.Episodes {
				for _, img := range ep.Images.URLs() {
					row := []string{t.Name, t.Category.String(), t.Link, ep.Name, ep.Link, img}
					if err := w.Write(row); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return []byte(sb.String()), nil
}

/* Returns the content type of the CSV formatter. */
func (CSVFormatter) ContentType() string {
	return "text/csv"
}
