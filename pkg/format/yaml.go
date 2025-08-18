package format

import (
	"gopkg.in/yaml.v3"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

type YAMLOutput struct {
	Total  int         `yaml:"total"`
	Titles []YAMLTitle `yaml:"titles"`
}

type YAMLTitle struct {
	Name     string        `yaml:"name"`
	Category string        `yaml:"category"`
	Url      string        `yaml:"url"`
	Episodes []YAMLEpisode `yaml:"episodes"`
	Images   []string      `yaml:"images,omitempty"`
}

type YAMLEpisode struct {
	Name   string   `yaml:"name"`
	Url    string   `yaml:"url"`
	Images []string `yaml:"images,omitempty"`
}

type YAMLFormatter struct{}

var yamlFmt = YAMLFormatter{}

/* Returns a YAML representation of titles `titles`. */
func (YAMLFormatter) Format(titles []*types.Title) ([]byte, error) {
	var yamlTitles []YAMLTitle
	for _, t := range titles {
		yamlTitle := YAMLTitle{
			Name:     t.Name,
			Category: t.Category.String(),
			Url:      t.Url,
			Images:   t.Images.URLs(),
		}
		for _, ep := range t.Episodes {
			yamlTitle.Episodes = append(yamlTitle.Episodes, YAMLEpisode{
				Name:   ep.Name,
				Url:    ep.Url,
				Images: ep.Images.URLs(),
			})
		}
		yamlTitles = append(yamlTitles, yamlTitle)
	}

	output := YAMLOutput{
		Total:  len(yamlTitles),
		Titles: yamlTitles,
	}

	return yaml.Marshal(output)
}

/* Returns the content type of the YAML formatter. */
func (YAMLFormatter) ContentType() string {
	return "application/x-yaml"
}
