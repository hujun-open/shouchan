package shouchan

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Action struct {
	name  string
	sconf SConfInt
	fset  *flag.FlagSet
}

/*
ActionConf represents a command with multiple actions, e.g. an zip file utility might have full command line structure:

- examplezip show -zipf <zip_filename> -fname <filename_inside_zip>
- examplezip zip -folder <foldername> -zipf <zip_filename>
- examplezip extract -zipf <zip_filename> -output <output_path>

"show/zip/extract" in this example represents different action, and each action requires different set of configuration/parameters,
ActionConf allows user to specify a map betweening an action(string) with *SConf[any]
*/
type ActionConf struct {
	//can't use *Sconf[any] here, see https://stackoverflow.com/questions/71399641/why-a-generic-cant-be-assigned-to-another-even-if-their-type-arguments-can
	list         map[string]*Action
	orderedList  []string //used to preserve the oder of action as it gets added, so that usage output is consistent
	loadedAction string
}

func newActionConf() *ActionConf {
	r := new(ActionConf)
	r.list = make(map[string]*Action)
	return r
}

type ActionConfig interface {
	DefaultCfgFilePath() string //return "" to be ignored
	ActionName() string
}

// newActionConfWithList creates ActionConf via list, key is the action, value is the a config struct implements ConfigWithDefCfgFilePath interface.
func newActionConfWithList(list []ActionConfig, options ...SconfOption[ActionConfig]) (*ActionConf, error) {
	r := newActionConf()
	var err error
	for _, action := range list {
		r.orderedList = append(r.orderedList, action.ActionName())
		r.list[action.ActionName()] = &Action{
			name: action.ActionName(),
			fset: flag.NewFlagSet(action.ActionName()+"-flagset", flag.ContinueOnError),
		}
		r.list[action.ActionName()].sconf, err = NewSConf(action, r.list[action.ActionName()].fset, options...)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// NewActionConfWithCMDLine creates ActionConf via list, key is the action, value is the a config struct implements ConfigWithDefCfgFilePath interface.
func NewActionConfWithCMDLine(list []ActionConfig, options ...SconfOption[ActionConfig]) (*ActionConf, error) {
	acnf, err := newActionConfWithList(list, options...)
	if err != nil {
		return nil, err
	}
	flag.CommandLine.Usage = acnf.PrintUsage
	return acnf, nil
}

var ErrEmptyArg = errors.New("empty argument list")

/*
Read loads configuration from args and/or config file
args format is following:
- if there is only one action: the same as SConf.Read
- otherwise, args[0] is the action, action is case-sensistive

return following errors:
- actionerr: error during parsing action, application should check & handle this error
- ferr&aerr: same sa SConf
*/
func (acnf *ActionConf) Read(args []string) (actionerr, ferr, aerr error) {
	if len(args) == 0 {
		return nil, nil, ErrEmptyArg
	}
	switch len(acnf.list) {
	case 0:
		return fmt.Errorf("no action defined"), nil, nil
	case 1:
		for action, scnf := range acnf.list {
			acnf.loadedAction = action
			ferr, aerr = scnf.sconf.Read(args)
			return
		}
	default:
		action := strings.TrimSpace(args[0])
		if action == "-?" {
			acnf.PrintUsage()
			return nil, nil, nil
		}
		if scnf, ok := acnf.list[action]; ok {
			acnf.loadedAction = action
			ferr, aerr = scnf.sconf.Read(args[1:])
			return
		}
		actionerr = fmt.Errorf("%v is not a valid action, use -? for list of actions", action)
		return
	}
	return fmt.Errorf("unusal error, should not happen"), nil, nil
}

func (acnf *ActionConf) ReadwithCMDLine() (actionerr, ferr, aerr error) {
	return acnf.Read(os.Args[1:])
}

// GetLoadedAction returns the loaded action, "" means not loaded
func (acnf *ActionConf) GetLoadedAction() string {
	return acnf.loadedAction
}

// GetLoadedConf returns loaded SConf, nil means not loaded
func (acnf *ActionConf) GetLoadedConf() any {
	if cnf, ok := acnf.list[acnf.loadedAction]; ok {
		return cnf.sconf.GetConfAny()
	}
	return nil
}

// PrintUsage print out the usage
func (acnf *ActionConf) PrintUsage() {
	fmt.Println("Usage: <action> [<parameters...>]")
	fmt.Printf("Actions: %s\n", strings.Join(acnf.orderedList, "|"))
	fmt.Println("Action specific usage:")
	for _, name := range acnf.orderedList {
		fmt.Printf("= %v\n", name)
		acnf.list[name].sconf.printUsage()
		fmt.Println()
	}
}
