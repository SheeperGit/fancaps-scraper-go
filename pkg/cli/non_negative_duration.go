package cli

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

/* A non-negative duration. */
type nnDuration time.Duration

/*
Returns a new non-negative duration value.
Panics if `val` is less than 0.
*/
func newNnDuration(val time.Duration, p *time.Duration) *nnDuration {
	if val < 0 {
		panic(fmt.Sprintf("default value for nnDuration must be non-negative (got: %v)", val))
	}

	*p = val
	return (*nnDuration)(p)
}

/*
Sets the nnDuration value `i` to a non-negative duration value derived from the string `s`.
Returns any errors encountered.
*/
func (i *nnDuration) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	if v < 0 {
		return fmt.Errorf("value must be non-negative (got: %v)", v)
	}
	*i = nnDuration(v)

	return nil
}

/* Returns the string representation of the nnDuration value `d`. */
func (d *nnDuration) String() string {
	return time.Duration(*d).String()
}

/* Returns a string representing the type of nnDuration `d`. */
func (d *nnDuration) Type() string {
	return "duration"
}

/* Registers a non-negative duration flag. */
func NnDurationVar(flagSet *pflag.FlagSet, p *time.Duration, name string, value time.Duration, usage string) {
	flagSet.Var(newNnDuration(value, p), name, fmt.Sprintf("%s (non-negative)", usage))
}
