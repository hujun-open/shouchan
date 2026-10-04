package shouchan

import (
	"fmt"
	"os"
	"strings"

	"github.com/hujun-open/extyaml"
	"github.com/hujun-open/myflags/v2"
	"github.com/spf13/cobra"
)

const (
	DefCfgFileFlagName = "cfgfromfile"
)

type SConfInt interface {
	read(args []string) (actionerr, ferr, aerr error)
	GetConfAny() any
	UsageStr(prefix string) string
}

// SConf represents a set of configurations as a struct
type SConf[X any] struct {
	conf               *X
	defaults           *X     // snapshot of def, restored at the start of every Read
	defConfFilePath    string //if this is empty, then there is no config file support
	parsedConfFilePath string
	Filler             *myflags.Filler
	fillerOptions      []myflags.FillerOption
	configFileFlagName string
	parsedActs         []string
	fillFlags          bool
}

type SconfOption[X any] func(ec *SConf[X])

// WithFillOptions specifies options used to create myflags.Filler
func WithFillOptions[X any](optlist []myflags.FillerOption) SconfOption[X] {
	return func(ec *SConf[X]) {
		ec.fillerOptions = optlist
	}
}

// WithFillFlags specifies whether to fill flags
func WithFillFlags[X any](fill bool) SconfOption[X] {
	return func(ec *SConf[X]) {
		ec.fillFlags = fill
	}
}

// WithDefaultConfigFilePath specifies default config file path, if it is empty, then there is no reading from config file
func WithDefaultConfigFilePath[X any](def string) SconfOption[X] {
	return func(ec *SConf[X]) {
		ec.defConfFilePath = def
	}
}

// WithConfigFileFlagName sepcifies the flag name of loading config file, default is defined by const DefCfgFileFlagName
func WithConfigFileFlagName[X any](name string) SconfOption[X] {
	return func(ec *SConf[X]) {
		ec.configFileFlagName = name
	}
}

// NewSConf returns a new SConf instance,
// def is a pointer to configruation struct with default value,
// defpath is the default configuration file path, it could be overriden by using command line arg "-f", could be "" means no default path
func NewSConf[X any](def *X, name, usage string, options ...SconfOption[X]) (*SConf[X], error) {
	if def == nil {
		return nil, fmt.Errorf("def is nil")
	}
	r := new(SConf[X])
	r.conf = def
	r.fillFlags = true
	r.configFileFlagName = DefCfgFileFlagName
	for _, o := range options {
		o(r)
	}
	if !r.fillFlags && r.defConfFilePath == "" {
		return nil, fmt.Errorf("default config file path is empty but also not instructed to fill flags")
	}
	r.Filler = myflags.NewFiller(name, usage, r.fillerOptions...)
	if r.fillFlags {
		err := r.Filler.Fill(r.conf)
		if err != nil {
			return nil, fmt.Errorf("failed to fill flagset, %w", err)
		}

	}
	if r.defConfFilePath != "" {
		if r.Filler.PersistentFlags().Lookup(r.configFileFlagName) != nil {
			return nil, fmt.Errorf("config file flag name %q collides with an existing flag", r.configFileFlagName)
		}
		r.Filler.PersistentFlags().StringVar(&r.parsedConfFilePath, r.configFileFlagName, r.defConfFilePath, "config file path")
	}
	r.defaults = new(X)
	cloneInto(reflectValue(r.defaults), reflectValue(r.conf))
	return r, nil
}

// configFileFromArgs returns the config file selected by args.
// Only the config-file flag is read. Every other argument is ignored, so a
// bad flag cannot skip loading the file or write into the config struct.
func (cnf *SConf[X]) configFileFromArgs(args []string) string {
	path := cnf.defConfFilePath
	if cnf.configFileFlagName == "" {
		return path
	}
	long := "--" + cnf.configFileFlagName
	prefix := long + "="
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		switch {
		case arg == long:
			if i+1 >= len(args) {
				return cnf.defConfFilePath
			}
			i++
			path = args[i]
		case strings.HasPrefix(arg, prefix):
			path = arg[len(prefix):]
		}
	}
	return path
}

func (cnf *SConf[X]) resetToDefaults() {
	if cnf.defaults == nil || cnf.conf == nil {
		return
	}
	cloneInto(reflectValue(cnf.conf), reflectValue(cnf.defaults))
}

// Read read configuration first from file, then from commandline args,
// commandline args will be read regardless if file read succeds,
// cmd is the command get executed,
// ferr is error of file reading, aerr is error of commandline args reading.
// if there is ferr and/or aerr, it could be treated as non-fatal failure thanks to mix&match and priority support.
func (cnf *SConf[X]) Read(args []string) (cmd *cobra.Command, ferr, aerr error) {
	cnf.resetToDefaults()
	if cnf.defConfFilePath != "" {
		path := cnf.configFileFromArgs(args)
		buf, err := os.ReadFile(path)
		if err != nil {
			ferr = fmt.Errorf("failed to open config file %v, %w", path, err)
		} else if err = cnf.UnmarshalYAML(buf); err != nil {
			ferr = fmt.Errorf("failed to decode %v as YAML, %w", path, err)
		}
	}
	cnf.Filler.SetArgs(args)
	cmd, aerr = cnf.Filler.ExecuteC()
	if aerr != nil {
		cmd = nil
		return
	}
	cnf.parsedActs = strings.Fields(cmd.CommandPath())[1:]
	return
}

// ReadCMDLine is same as Read, expcept the args is os.Args[1:]
func (cnf *SConf[X]) ReadwithCMDLine() (cmd *cobra.Command, ferr, aerr error) {
	return cnf.Read(os.Args[1:])
}

// MarshalYAML marshal config value into YAML
func (cnf *SConf[X]) MarshalYAML() ([]byte, error) {
	return extyaml.MarshalExt(cnf.conf)
}

// UnmarshalYAML unmrshal YAML encoded buf into config value.
// Fields present in buf replace the current value. Omitted fields are left as they are.
func (cnf *SConf[X]) UnmarshalYAML(buf []byte) error {
	if len(strings.TrimSpace(string(buf))) == 0 {
		return nil
	}
	fresh := new(X)
	if err := extyaml.UnmarshalExt(buf, fresh); err != nil {
		return err
	}
	return mergeYAML(reflectValue(cnf.conf), reflectValue(fresh), buf)
}

// GetConf returns config value
func (cnf *SConf[X]) GetConf() *X {
	return cnf.conf
}

func (cnf *SConf[X]) GetConfAny() any {
	return cnf.conf
}
