package cli

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

/* A single optional enum value type. */
type enumOptValue[T cmp.Ordered] struct {
	value *T           // Value.
	meta  *enumMeta[T] // Enum metadata.
}

/*
Returns a new optional enum value.
Panics if `value` is not present in `enumToVal`.
*/
func newEnumOptValue[T cmp.Ordered](value T, p *T, enumToVal map[string]T) *enumOptValue[T] {
	meta := newEnumMeta([]T{value}, enumToVal)

	/* Validate default value. */
	if _, ok := meta.valToEnum[value]; !ok {
		panic(fmt.Sprintf("default value %v not present in enums (valid values: %s)", value, meta.validEnums))
	}

	*p = value

	return &enumOptValue[T]{
		value: p,
		meta:  meta,
	}
}

/*
Sets the optional enum `e` to a non-empty value from an enum string `s`.
Returns any errors encountered.
*/
func (e *enumOptValue[T]) Set(s string) error {
	s = strings.ToLower(strings.TrimSpace(s))
	val, ok := e.meta.enumToVal[s]
	if !ok {
		return fmt.Errorf("invalid value %q; must be one of: %s", s, e.meta.validEnums)
	}
	*e.value = val

	return nil
}

/* Returns the string representation of the optional enum `e`. */
func (e *enumOptValue[T]) String() string {
	if name, ok := e.meta.valToEnum[*e.value]; ok {
		return name
	}

	return ""
}

/* Returns a string representing the type of optional enum `e`. */
func (e *enumOptValue[T]) Type() string {
	return "string"
}

/* Register an enum flag with an optional argument. */
func EnumOptVarP[T cmp.Ordered](flagSet *pflag.FlagSet, p *T, name, shorthand string, value T, enumToVal map[string]T, usage string) {
	ev := newEnumOptValue(value, p, enumToVal)
	flagSet.VarP(ev, name, shorthand, fmt.Sprintf("%s (allowed: %s)", usage, ev.meta.validEnums))

	/* Set a default value when specified with no value. */
	flagSet.Lookup(name).NoOptDefVal = ev.meta.valToEnum[value]
}
