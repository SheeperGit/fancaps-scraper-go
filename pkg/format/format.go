package format

import (
	"fmt"
	"os"
	"strings"

	"sheeper.com/fancaps-scraper-go/pkg/types"
	"sheeper.com/fancaps-scraper-go/pkg/ui"
)

/* Formats titles. */
type Formatter interface {
	Format(titles []*types.Title) ([]byte, error) // Defines how the titles `titles` are formatted.
	ContentType() string                          // Content type.
}

/* Available formatters. */
var formatters = []Formatter{
	jsonFmt,
	yamlFmt,
	csvFmt,
}

/* Maps content type to its corresponding formatter. */
var formatMap = func() map[string]Formatter {
	m := make(map[string]Formatter, len(formatters))
	for _, f := range formatters {
		m[f.ContentType()] = f
	}

	return m
}()

/* Available, comma-separated content types. */
var contentTypes = func() string {
	cts := make([]string, len(formatters))
	for i, f := range formatters {
		cts[i] = f.ContentType()
	}

	return strings.Join(cts, ", ")
}()

/* Prints titles `titles` in the format of the content type `contentType`. */
func OutputFormat(titles []*types.Title, contentType string) {
	f, ok := formatMap[contentType]
	if !ok {
		fmt.Fprintf(os.Stderr, ui.ErrStyle.Render("unknown content type `%s`")+"\n"+
			ui.ErrStyle.Render("valid content types: ")+ui.HighlightStyle.Render("%s"),
			contentType, contentTypes)
		os.Exit(1)
	}

	output, err := f.Format(titles)
	if err != nil {
		fmt.Fprintf(os.Stderr, ui.ErrStyle.Render("format error: %v")+"\n", err)
		os.Exit(1)
	}

	fmt.Print(string(output))
}
