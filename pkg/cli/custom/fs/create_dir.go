package cli

import (
	"fmt"

	"github.com/spf13/pflag"
	"sheeper.com/fancaps-scraper-go/pkg/fsutil"
)

/* A directory with existing parent directories. */
type createDirValue string

/*
Returns a new directory value.
Panics if `val` is does not have existing parent directories.
*/
func newCreateDirValue(val string, p *string) *createDirValue {
	if !fsutil.ParentDirsExist(val) {
		panic(fmt.Sprintf("default value for createDir must have existing parent directories (got: %v)", val))
	}

	*p = val
	return (*createDirValue)(p)
}

/*
Sets the directory value `d` to a directory with existing parent directories `s`.
Returns any errors encountered.
*/
func (d *createDirValue) Set(s string) error {
	if !fsutil.ParentDirsExist(s) {
		return fmt.Errorf("invalid output directory %s; make sure the parent directories exist", s)
	}
	*d = createDirValue(s)

	return nil
}

/* Returns the filepath of the directory value `d`. */
func (d *createDirValue) String() string {
	if d == nil {
		return ""
	}

	return string(*d)
}

/* Returns a string representing the type of directory value `d`. */
func (d *createDirValue) Type() string {
	return "string"
}

/* Registers a create directory flag. */
func CreateDirVarP(flagSet *pflag.FlagSet, p *string, name, shorthand, value, usage string) {
	flagSet.VarP(newCreateDirValue(value, p), name, shorthand, fmt.Sprintf("%s (parent directories must exist)", usage))
}
