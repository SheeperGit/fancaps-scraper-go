package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/pflag"
)

/* A strictly positive uint8. (1-255) */
type puint8Value uint8

/*
Returns a new puint8 value.
Panics if `val` is 0.
*/
func newPuint8Value(val uint8, p *uint8) *puint8Value {
	if val == 0 {
		panic(fmt.Sprintf("default value for puint8 must be strictly positive (got: %d)", val))
	}

	*p = val
	return (*puint8Value)(p)
}

/*
Sets the puint8 value `i` to a strictly positive uint8 value derived from the string `s`.
Returns any errors encountered.
*/
func (i *puint8Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return err
	}
	if v == 0 {
		return fmt.Errorf("invalid value %q; must be strictly positive (1-255)", s)
	}
	*i = puint8Value(v)

	return nil
}

/* Returns the string representation of the puint8 value `i`. */
func (i *puint8Value) String() string {
	return strconv.FormatUint(uint64(*i), 10)
}

/* Returns a string representing the type of puint8 `i`. */
func (i *puint8Value) Type() string {
	return "uint8"
}

/* Registers a strictly positive uint8 flag. */
func Puint8VarP(flagSet *pflag.FlagSet, p *uint8, name, shorthand string, value uint8, usage string) {
	flagSet.VarP(newPuint8Value(value, p), name, shorthand, fmt.Sprintf("%s (1-255)", usage))
}
