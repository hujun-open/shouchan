package shouchan

import (
	"fmt"
	"os"
	"reflect"

	"github.com/hujun-open/extyaml"
	"github.com/hujun-open/myflags"
)

const (
	defCfgFileFlagName = "-cfgfromfile"
)

type SConfInt interface {
	read(args []string) (actionerr, ferr, aerr error)
	GetConfAny() any
	UsageStr(prefix string) string
}

// SConf represents a set of configurations as a struct
type SConf[T any] struct {
	conf               T
	defConfFilePath    string //if this is empty, then there is no config file support
	filler             *myflags.Filler
	configFileFlagName string
	parsedActs         []string
}

type SconfOption[T any] func(ec *SConf[T])

func WithDefaultConfigFilePath[T any](def string) SconfOption[T] {
	return func(ec *SConf[T]) {
		ec.defConfFilePath = def
	}
}

func WithConfigFileFlagName[T any](name string) SconfOption[T] {
	return func(ec *SConf[T]) {
		ec.configFileFlagName = name
	}
}

// NewSConf returns a new SConf instance,
// def is a pointer to configruation struct with default value,
// defpath is the default configuration file path, it could be overriden by using command line arg "-f", could be "" means no default path
func NewSConf[T any](def T, name, usage string, options ...SconfOption[T]) (*SConf[T], error) {
	if reflect.TypeOf(def).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("def is not a ptr")
	}
	r := new(SConf[T])
	r.conf = def
	r.configFileFlagName = defCfgFileFlagName
	r.filler = myflags.NewFiller(name, usage)
	for _, o := range options {
		o(r)
	}
	err := r.filler.Fill(r.conf)
	if err != nil {
		return nil, fmt.Errorf("failed to fill flagset, %w", err)
	}
	r.filler.GetFlagset().Usage = r.PrintUsage
	return r, nil
}

func (cnf *SConf[T]) getConfFilePath(args []string) (string, []string) {
	for i, arg := range args {
		if arg == cnf.configFileFlagName && i < len(args)-1 {
			fpstr := args[i+1]
			return fpstr, append(args[:i], args[i+2:]...)
		}
	}
	return "", args
}

func (cnf *SConf[T]) read(args []string) (actionerr, ferr, aerr error) {
	ferr, aerr = cnf.Read(args)
	return nil, ferr, aerr
}

// Read read configuration first from file, then flagset from args,
// flagset will be read regardless if file read succeds,
// ferr is error of file reading, aerr is error of flagset reading.
// if there is ferr and/or aerr, it could be treated as non-fatal failure thanks to mix&match and priority support.
func (cnf *SConf[T]) Read(args []string) (ferr, aerr error) {
	var buf []byte
	var fpath string
	newargs := args
	if cnf.defConfFilePath != "" {
		fpath, newargs = cnf.getConfFilePath(args)
		if fpath == "" {
			fpath = cnf.defConfFilePath
		}
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
	cnf.parsedActs, aerr = cnf.filler.ParseArgs(newargs)
	return
}

func (cnf *SConf[T]) PrintUsage() {
	fmt.Print(cnf.UsageStr(""))
}

func (cnf *SConf[T]) UsageStr(prefix string) string {
	return cnf.filler.UsageStr("") + fmt.Sprintf("\n  %v: load configuration from the specified file\n        default:%v\n",
		cnf.configFileFlagName, cnf.defConfFilePath)
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
