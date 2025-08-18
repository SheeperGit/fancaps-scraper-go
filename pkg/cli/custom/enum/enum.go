package cli

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

/* A single enum value type. */
type enumValue[T cmp.Ordered] struct {
	value *T           // Value.
	meta  *enumMeta[T] // Enum metadata.
}

/*
Returns a new enum value.
Panics if `value` is not present in `enumToVal`.
*/
func newEnumValue[T cmp.Ordered](value T, p *T, enumToVal map[string]T) *enumValue[T] {
	meta := newEnumMeta([]T{value}, enumToVal)

	*p = value

	return &enumValue[T]{
		value: p,
		meta:  meta,
	}
}

/*
Sets the enum `e` to a non-empty value from an enum string `s`.
Returns any errors encountered.
*/
func (e *enumValue[T]) Set(s string) error {
	s = strings.ToLower(strings.TrimSpace(s))

	val, ok := e.meta.enumToVal[s]
	if !ok {
		return fmt.Errorf("invalid value %q; must be one of: %s", s, e.meta.validEnums)
	}

	*e.value = val
	return nil
}

/* Returns the string representation of the enum `e`. */
func (e *enumValue[T]) String() string {
	if e.value == nil {
		return ""
	}

	if name, ok := e.meta.valToEnum[*e.value]; ok {
		return name
	}

	return ""
}

/* Returns a string representing the type of enum `e`. */
func (e *enumValue[T]) Type() string {
	return "string"
}

/* Register an enum flag. */
func EnumVar[T cmp.Ordered](flagSet *pflag.FlagSet, p *T, name string, value T, enumToVal map[string]T, usage string) {
	ev := newEnumValue(value, p, enumToVal)
	flagSet.Var(ev, name, fmt.Sprintf("%s (allowed: %s)", usage, ev.meta.validEnums))
}
