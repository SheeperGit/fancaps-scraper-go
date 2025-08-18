package format

import (
	"strconv"
	"strings"

	"sheeper.com/fancaps-scraper-go/pkg/types"
)

const (
	titleSpacing   = "  "
	episodeSpacing = titleSpacing + titleSpacing
)

type DEFFormatter struct{}

var defFmt = DEFFormatter{}

/* Returns the default representation of titles `titles`. */
func (DEFFormatter) Format(titles []*types.Title) ([]byte, error) {
	/* Write image strings `images` to `sb` prefixed with `prefix`. */
	writeImages := func(sb *strings.Builder, prefix string, images []string) {
		if len(images) == 0 {
			return
		}

		sb.WriteString(prefix + "images:\n")
		for _, img := range images {
			sb.WriteString(prefix + titleSpacing + img + "\n")
		}
	}

	var sb strings.Builder

	sb.WriteString("total: " + strconv.Itoa(len(titles)) + "\n")

	sb.WriteString("titles:\n")
	for _, t := range titles {
		sb.WriteString(titleSpacing + t.Name + " [" + t.Category.String() + "]: " + t.Url + "\n")
		writeImages(&sb, titleSpacing, t.Images.URLs())

		sb.WriteString(titleSpacing + "episodes:\n")
		for _, ep := range t.Episodes {
			sb.WriteString(episodeSpacing + ep.Name + ": " + ep.Url + "\n")
			writeImages(&sb, episodeSpacing, ep.Images.URLs())
		}
	}

	return []byte(sb.String()), nil
}

/* Returns the content type of the default formatter. */
func (DEFFormatter) ContentType() string {
	return "text/plain; charset=utf-8"
}
