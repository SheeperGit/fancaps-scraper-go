package cli

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

/* A slice of enums. */
type enumSliceValue[T cmp.Ordered] struct {
	value *[]T         // Values.
	meta  *enumMeta[T] // Enum metadata.
}

/*
Returns a new enum slice.
Panics if `value` contains an element not present in the values of `enumToVal`.
*/
func newEnumSliceValue[T cmp.Ordered](value []T, p *[]T, enumToVal map[string]T) *enumSliceValue[T] {
	meta := newEnumMeta(value, enumToVal)

	*p = value

	return &enumSliceValue[T]{
		value: p,
		meta:  meta,
	}
}

/*
Sets the enum slice `e` to a non-empty, unique, sorted list of
values from an enum string `s`.
Returns any errors encountered.
*/
func (e *enumSliceValue[T]) Set(s string) error {
	parts := strings.Split(s, ",")
	seen := map[T]bool{}
	unique := []T{}

	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))

		val, ok := e.meta.enumToVal[p]
		if !ok {
			return fmt.Errorf("invalid value %q; must be one of: %s", p, e.meta.validEnums)
		}
		if !seen[val] {
			seen[val] = true
			unique = append(unique, val)
		}
	}
	slices.Sort(unique)

	*e.value = unique
	return nil
}

/* Returns a comma-separated string of enums from an enum slice `e`. */
func (e *enumSliceValue[T]) String() string {
	if e.value == nil || *e.value == nil {
		return ""
	}

	enums := make([]string, len(*e.value))
	for i, v := range *e.value {
		enums[i] = e.meta.valToEnum[v]
	}

	return strings.Join(enums, ", ")
}

/* Returns a string representing the type of enum slice `e`. */
func (e *enumSliceValue[T]) Type() string {
	return "strings"
}

/* Registers an enum slice flag. */
func EnumSliceVarP[T cmp.Ordered](flagSet *pflag.FlagSet, p *[]T, name, shorthand string, value []T, enumToVal map[string]T, usage string) {
	ev := newEnumSliceValue(value, p, enumToVal)
	flagSet.VarP(ev, name, shorthand, fmt.Sprintf("%s (allowed: %s)", usage, ev.meta.validEnums))
}
