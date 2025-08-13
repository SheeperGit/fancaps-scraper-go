package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/pflag"
	"sheeper.com/fancaps-scraper-go/pkg/types"
)

/* A slice of categories derived from enums. */
type enumCategorySliceValue struct {
	value *[]types.Category
}

var enumToCategory = map[string]types.Category{
	"anime":  types.CategoryAnime,
	"tv":     types.CategoryTV,
	"movies": types.CategoryMovie,
}

var categoryToEnum = func(m map[string]types.Category) map[types.Category]string {
	reverse := make(map[types.Category]string, len(m))
	for k, v := range m {
		reverse[v] = k
	}
	return reverse
}(enumToCategory)

var validEnums = func(etoC map[string]types.Category) string {
	cats := slices.Collect(maps.Values(etoC))
	slices.Sort(cats) // Sort by Category enum order.

	names := make([]string, len(cats))
	for i, cat := range cats {
		names[i] = categoryToEnum[cat]
	}

	return strings.Join(names, ", ")
}(enumToCategory)

/*
Sets the category slice `e` to a non-empty, unique, sorted list of
categories from an enum string `s`.
Returns any errors encountered.
*/
func (e *enumCategorySliceValue) Set(s string) error {
	parts := strings.Split(s, ",")
	seen := map[types.Category]bool{}
	unique := []types.Category{}

	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		cat, ok := enumToCategory[p]
		if !ok {
			return fmt.Errorf("invalid value %q; must be one of: %s",
				p, validEnums)
		}
		if !seen[cat] {
			seen[cat] = true
			unique = append(unique, cat)
		}
	}
	slices.Sort(unique) // Sort by Category enum order.
	*e.value = unique

	return nil
}

/* Returns a comma-separated string of categories from a category slice `e`. */
func (e *enumCategorySliceValue) String() string {
	if e.value == nil || *e.value == nil {
		return ""
	}

	names := make([]string, len(*e.value))
	for i, cat := range *e.value {
		names[i] = categoryToEnum[cat]
	}

	return strings.Join(names, ", ")
}

/* Returns a string representing the type of category slice `e`. */
func (e *enumCategorySliceValue) Type() string {
	return "categories"
}

/* Registers a category slice flag. */
func CategorySliceVarP(flagSet *pflag.FlagSet, p *[]types.Category, name, shorthand string, value []types.Category, usage string) {
	*p = value
	flagSet.VarP(&enumCategorySliceValue{
		value: p,
	}, name, shorthand, fmt.Sprintf("%s (allowed: %s)", usage, validEnums))
}
