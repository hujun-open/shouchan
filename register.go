package shouchan

import (
	"reflect"

	"github.com/hujun-open/extyaml"
	"github.com/hujun-open/myflags"
)

// FromStr is the function convert a string into a instance of to-be-supported-type
type FromStr func(s string) (any, error)

// ToStr is the function convert a instance of to-be-supported-type into string
type ToStr func(in any) (string, error)

type flagConverter struct {
	from FromStr
	to   ToStr
}

func (fc *flagConverter) ToStr(in any, tag reflect.StructTag) string {
	r, _ := fc.to(in)
	return r
}

func (fc *flagConverter) FromStr(s string, tag reflect.StructTag) (any, error) {
	return fc.from(s)
}

// Register type T with provided to and from functions
func Register[T any](to ToStr, from FromStr) {
	extyaml.RegisterExt[T](extyaml.ToStr(to), extyaml.FromStr(from))
	myflags.Register[T](&flagConverter{
		to:   to,
		from: from,
	})
}
