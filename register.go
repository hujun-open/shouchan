package shouchan

import (
	"reflect"

	"github.com/hujun-open/extyaml"
	"github.com/itzg/go-flagsfiller"
)

// FromStr is the function convert a string into a instance of to-be-supported-type
type FromStr func(s string) (any, error)

// ToStr is the function convert a instance of to-be-supported-type into string
type ToStr func(in any) (string, error)

// Register type T with provided to and from functions
func Register[T any](to ToStr, from FromStr) {
	extyaml.RegisterExt[T](extyaml.ToStr(to), extyaml.FromStr(from))
	flagCvt := func(s string, tag reflect.StructTag) (T, error) {
		r, err := from(s)
		return r.(T), err
	}
	flagsfiller.RegisterSimpleType(flagCvt)
}
