package cli

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

/* Shared metadata for enum types. */
type enumMeta[T cmp.Ordered] struct {
	enumToVal  map[string]T // Custom enums to value.
	valToEnum  map[T]string // Value to custom enums.
	validEnums string       // Valid comma-separated enums.
}

/*
Returns a new enum metadata struct for enums.
If the list of values `value` contains a invalid value elements, this function panics.
*/
func newEnumMeta[T cmp.Ordered](value []T, enumToVal map[string]T) *enumMeta[T] {
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

		return "[" + strings.Join(names, "|") + "]"
	}()

	/* Validate default values. */
	for _, v := range value {
		if _, ok := valToEnum[v]; !ok {
			panic(fmt.Sprintf("default value %v not present in enums (valid values: %s)", v, validEnums))
		}
	}

	return &enumMeta[T]{
		enumToVal:  enumToVal,
		valToEnum:  valToEnum,
		validEnums: validEnums,
	}
}
