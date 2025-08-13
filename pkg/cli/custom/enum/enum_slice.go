package cli

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

/* A slice of T's derived from enums. */
type enumSliceValue[T cmp.Ordered] struct {
	value      *[]T         // Values.
	enumToVal  map[string]T // Custom enums to value.
	valToEnum  map[T]string // Value to custom enums.
	validEnums string       // Valid comma-separated enums.
}

/*
Returns a new enum slice.
Panics if `T` contains an element not present in the values of `enumToVal`.
*/
func newEnumSliceValue[T cmp.Ordered](value []T, p *[]T, enumToVal map[string]T) *enumSliceValue[T] {
	/* Returns a reverse map of a map from enum to value `m`. */
	valToEnum := func(m map[string]T) map[T]string {
		reverse := make(map[T]string, len(m))
		for k, v := range m {
			reverse[v] = k
		}

		return reverse
	}(enumToVal)

	/* Returns a sorted (according to enum order) string list of valid enums. */
	validEnums := func() string {
		vals := slices.Collect(maps.Values(enumToVal))
		slices.Sort(vals)

		names := make([]string, len(vals))
		for i, v := range vals {
			names[i] = valToEnum[v]
		}

		return strings.Join(names, ", ")
	}()

	/* Validate default values. */
	for _, v := range value {
		if _, ok := valToEnum[v]; !ok {
			panic(fmt.Sprintf("default value %v not present in enums (valid values: %s)", v, validEnums))
		}
	}

	*p = value

	return &enumSliceValue[T]{
		value:      p,
		enumToVal:  enumToVal,
		valToEnum:  valToEnum,
		validEnums: validEnums,
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
		val, ok := e.enumToVal[p]
		if !ok {
			return fmt.Errorf("invalid value %q; must be one of: %s", p, e.validEnums)
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
		enums[i] = e.valToEnum[v]
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

	flagSet.VarP(ev, name, shorthand, fmt.Sprintf("%s (allowed: %s)", usage, ev.validEnums))
}
