package shouchan

import (
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
	loadedAction string
}

func newActionConf() *ActionConf {
	r := new(ActionConf)
	r.list = make(map[string]*Action)
	return r
}

type ConfigWithDefCfgFilePath interface {
	DefaultCfgFilePath() string //return "" to be ignored
}

// newActionConfWithMap creates ActionConf via list, key is the action, value is the a config struct implements ConfigWithDefCfgFilePath interface.
func newActionConfWithMap(list map[string]ConfigWithDefCfgFilePath) (*ActionConf, error) {
	r := newActionConf()
	var err error
	for action, v := range list {
		r.list[action] = &Action{
			name: action,
			fset: flag.NewFlagSet(action+"-flagset", flag.ContinueOnError),
		}
		r.list[action].sconf, err = NewSConf(v, "", r.list[action].fset)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// NewActionConfWithCMDLine creates ActionConf via list, key is the action, value is the a config struct implements ConfigWithDefCfgFilePath interface.
func NewActionConfWithCMDLine(list map[string]ConfigWithDefCfgFilePath) (*ActionConf, error) {
	acnf, err := newActionConfWithMap(list)
	if err != nil {
		return nil, err
	}
	flag.CommandLine.Usage = acnf.printUsage
	return acnf, nil
}

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
		return nil, nil, fmt.Errorf("empty argument list")
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
			acnf.printUsage()
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

func (acnf *ActionConf) printUsage() {
	fmt.Println("Usage: <action> [<parameters...>]")
	nameList := []string{}
	for name := range acnf.list {
		nameList = append(nameList, name)
	}
	fmt.Printf("Actions: %s\n", strings.Join(nameList, "|"))
	fmt.Println("Action specific usage:")
	for _, name := range nameList {
		fmt.Printf("= %v\n", name)
		acnf.list[name].sconf.printUsage()
		fmt.Println()
	}
}
