package cli

import (
	"fmt"

	"github.com/spf13/pflag"
	"sheeper.com/fancaps-scraper-go/pkg/fsutil"
)

/* A directory with existing parent directories. */
type createDirValue struct {
	value *string
}

/*
Sets the directory value `d` to a directory with existing parent directories `s`.
Returns any errors encountered.
*/
func (d *createDirValue) Set(s string) error {
	if !fsutil.ParentDirsExist(s) {
		return fmt.Errorf("invalid output directory %s; make sure the parent directories exists", s)
	}
	*d.value = s

	return nil
}

/* Returns the filepath of the directory value `d`. */
func (d *createDirValue) String() string {
	if d.value == nil {
		return ""
	}

	return *d.value
}

/* Returns a string representing the type of directory value `d`. */
func (d *createDirValue) Type() string {
	return "directory"
}

/* Registers a create directory flag. */
func CreateDirVarP(flagSet *pflag.FlagSet, p *string, name, shorthand, value, usage string) {
	*p = value

	flagSet.VarP(&createDirValue{value: p}, name, shorthand, usage)
}
