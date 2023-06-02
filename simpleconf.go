package shouchan

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/hujun-open/extyaml"
	"github.com/itzg/go-flagsfiller"
)

const (
	confArgName = "-f"
)

type SConfInt interface {
	Read(args []string) (ferr, aerr error)
	GetConfAny() any
	printUsage()
}

// SConf represents a set of configurations as a struct
type SConf[T any] struct {
	conf         T
	confFilePath string
	filler       *flagsfiller.FlagSetFiller
	fset         *flag.FlagSet
}

// NewSConf returns a new SConf instance,
// def is a pointer to configruation struct with default value,
// defpath is the default configuration file path, it could be overriden by using command line arg "-f", could be "" means no default path
// fset is the flagset to bind
func NewSConf[T any](def T, defpath string, fset *flag.FlagSet) (*SConf[T], error) {
	if reflect.TypeOf(def).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("def is not a ptr")
	}
	r := new(SConf[T])
	r.conf = def
	r.confFilePath = defpath
	r.fset = fset
	r.filler = flagsfiller.New(flagsfiller.WithFieldRenamer(
		strings.ToLower))
	err := r.filler.Fill(fset, r.conf)
	if err != nil {
		return nil, fmt.Errorf("failed to fill flagset, %w", err)
	}
	r.fset.Usage = r.printUsage
	return r, nil
}

// NewSConfCMDLine is same as NewSConf, just use flag.CommandLine as the flagset
func NewSConfCMDLine[T any](def T, defpath string) (*SConf[T], error) {
	return NewSConf(def, defpath, flag.CommandLine)
}

func getConfFilePath(args []string) (string, []string) {
	for i, arg := range args {
		if arg == confArgName && i < len(args)-1 {
			fpstr := args[i+1]
			return fpstr, append(args[:i], args[i+2:]...)
		}
	}
	return "", args
}

// Read read configuration first from file, then flagset from args,
// flagset will be read regardless if file read succeds,
// ferr is error of file reading, aerr is error of flagset reading.
// if there is ferr and/or aerr, it could be treated as non-fatal failure thanks to mix&match and priority support.
func (cnf *SConf[T]) Read(args []string) (ferr, aerr error) {
	var buf []byte
	fpath := cnf.confFilePath
	newargs := args
	if fpath == "" {
		fpath, newargs = getConfFilePath(args)
	}
	if fpath != "" {
		buf, ferr = os.ReadFile(fpath)
		if ferr != nil {
			ferr = fmt.Errorf("failed to open config file %v, %w", fpath, ferr)
		} else {
			ferr = cnf.UnmarshalYAML(buf)
			if ferr != nil {
				ferr = fmt.Errorf("failed to decode %v as YAML, %w", fpath, ferr)
			}
		}
	}
	aerr = cnf.fset.Parse(newargs)
	return
}

func (cnf *SConf[T]) printUsage() {
	indent := "  "
	fmt.Println("Usage:")
	fmt.Printf("%v-f <filepath> : read from config file <filepath>\n", indent)
	cnf.fset.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%v-%v <%v> : %v\n", indent, f.Name,
			reflect.Indirect(reflect.ValueOf(f.Value)).Kind(),
			f.Usage)
		if f.DefValue != "" {
			fmt.Printf("%v\tdefault:%v\n", indent, f.DefValue)
		}
	})
}

// ReadCMDLine is same as Read, expcept the args is os.Args[1:]
func (cnf *SConf[T]) ReadwithCMDLine() (ferr, aerr error) {
	return cnf.Read(os.Args[1:])
}

// MarshalYAML marshal config value into YAML
func (cnf *SConf[T]) MarshalYAML() ([]byte, error) {
	return extyaml.MarshalExt(cnf.conf)
}

// UnmarshalYAML unmrshal YAML encoded buf into config value
func (cnf *SConf[T]) UnmarshalYAML(buf []byte) error {
	return extyaml.UnmarshalExt(buf, cnf.conf)
}

// GetConf returns config value
func (cnf *SConf[T]) GetConf() T {
	return cnf.conf
}

func (cnf *SConf[T]) GetConfAny() any {
	return cnf.conf
}
