package shouchan

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/hujun-open/cobra"
	"github.com/hujun-open/extyaml"
	"github.com/hujun-open/myflags/v2"
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
	if reflect.TypeOf(def).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("def is not a ptr")
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
		r.Filler.PersistentFlags().StringVar(&r.parsedConfFilePath, r.configFileFlagName, r.defConfFilePath, "config file path")
	}
	// r.filler.GetFlagset().Usage = r.PrintUsage
	return r, nil
}

// disableCommands recursively disables all execution hooks for a command and its subcommands.
// also disable the help output
func disableCommands(cmd *cobra.Command) {
	cmd.Run = myflags.DefRunMethod
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	cmd.PreRun = myflags.DefRunMethod
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
	cmd.PostRun = myflags.DefRunMethod
	cmd.PostRunE = func(cmd *cobra.Command, args []string) error { return nil }
	cmd.SilenceUsage = true
	cmd.SetHelpCommand(nil)
	cmd.SetHelpFunc(func(*cobra.Command, []string) {})

	for _, subCmd := range cmd.Commands() {
		disableCommands(subCmd)
	}
}

// clone create a new SConf instnace that inherit from cnf, but with new .conf filled,
// so that it could be used by Read to get the configFileflag value without impacting the cnf.conf
func (cnf *SConf[X]) clone() *SConf[X] {
	newone := new(SConf[X])
	*newone = *cnf
	newone.conf = new(X)
	*newone.conf = *cnf.conf

	newone.Filler = myflags.NewFiller(cnf.Filler.Name(), cnf.Filler.UsageString(), newone.fillerOptions...)

	err := newone.Filler.Fill(newone.conf)
	if err != nil {
		panic(err)
	}

	if newone.defConfFilePath != "" {
		newone.Filler.PersistentFlags().StringVar(&newone.parsedConfFilePath, newone.configFileFlagName, newone.defConfFilePath, "config file path")
	}
	disableCommands(newone.Filler.Command)
	return newone

}

// Read read configuration first from file, then from commandline args,
// commandline args will be read regardless if file read succeds,
// cmd is the command get executed,
// ferr is error of file reading, aerr is error of commandline args reading.
// if there is ferr and/or aerr, it could be treated as non-fatal failure thanks to mix&match and priority support.
func (cnf *SConf[X]) Read(args []string) (cmd *cobra.Command, ferr, aerr error) {
	var buf []byte
	newargs := args
	if cnf.defConfFilePath != "" {
		tempcnf := cnf.clone()
		tempcnf.Filler.SetArgs(args)
		if nerr := tempcnf.Filler.Execute(); nerr == nil {
			buf, ferr = os.ReadFile(tempcnf.parsedConfFilePath)
			if ferr != nil {
				ferr = fmt.Errorf("failed to open config file %v, %w", tempcnf.parsedConfFilePath, ferr)
			} else {
				ferr = cnf.UnmarshalYAML(buf)
				if ferr != nil {
					ferr = fmt.Errorf("failed to decode %v as YAML, %w", tempcnf.parsedConfFilePath, ferr)
				}
			}
		}
	}
	cnf.Filler.SetArgs(newargs)
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

// UnmarshalYAML unmrshal YAML encoded buf into config value
func (cnf *SConf[X]) UnmarshalYAML(buf []byte) error {
	return extyaml.UnmarshalExt(buf, cnf.conf)
}

// GetConf returns config value
func (cnf *SConf[X]) GetConf() *X {
	return cnf.conf
}

func (cnf *SConf[X]) GetConfAny() any {
	return cnf.conf
}
