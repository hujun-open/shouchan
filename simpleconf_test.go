package shouchan

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/hujun-open/myflags/v2"
	_ "github.com/hujun-open/myflags/v2/types"
)

type company struct {
	Name string
}

type testStruct struct {
	Name, Addr string
	Employer   company
	JoinTime   time.Time
	NumList    []int
	Act        struct {
		NetName string
	} `action:"" usage:"testaction"`
}

func (t testStruct) isEqual(peer testStruct) bool {
	if t.Name == peer.Name && t.Addr == peer.Addr {
		if t.Employer.Name == peer.Employer.Name {
			if t.JoinTime.Equal(peer.JoinTime) {
				if t.Act.NetName == peer.Act.NetName {
					return true
				}
			}
		}
	}
	return false
}

const (
	defpath string = "./testdata/test.yaml"
)

type testSetup struct {
	def        testStruct
	fpath      string
	args       []string
	result     testStruct
	expectFail bool
	dontDoFlag bool
}

func doTest(t *testing.T, setup testSetup) error {

	options := []SconfOption[testStruct]{}
	options = append(options, WithDefaultConfigFilePath[testStruct](setup.fpath))
	options = append(options, WithFillFlags[testStruct](!setup.dontDoFlag))
	if setup.expectFail {
		options = append(options, WithFillOptions[testStruct]([]myflags.FillerOption{
			myflags.WithFlagErrHandling(flag.ContinueOnError),
		}))
	}
	cnf, err := NewSConf(&setup.def, "test", "golangdevtest", options...)
	if err != nil {
		return err
	}
	_, ferr, aerr := cnf.Read(setup.args)
	t.Logf("ferr is %v, aerr is %v", ferr, aerr)
	t.Logf("result conf is %+v", cnf.GetConf())
	if !cnf.GetConf().isEqual(setup.result) {

		return fmt.Errorf("actual result is %+v, different from expected result %+v", cnf.GetConf(), setup.result)
	}
	return nil

}
func TestSconf(t *testing.T) {
	defCnf := testStruct{
		Name:     "defName",
		Addr:     "defAddr",
		Employer: company{Name: "defCom"},
	}
	defCnf.JoinTime, _ = time.Parse(time.DateTime, "1999-01-02 03:04:05")
	caseList := []testSetup{
		{ // case 0, result should be value from file
			def:   defCnf,
			fpath: defpath,
			args:  []string{},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromFile",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"},
			},
		},
		{ // case 1, specify config file in args, result should be value from file
			def:   defCnf,
			fpath: "somenonexistingfilepath",
			args:  []string{"--" + DefCfgFileFlagName, defpath},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromFile",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"}},
		},
		{
			// case 2,no file, no args, result should be default
			def:   defCnf,
			fpath: "",
			args:  []string{},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "defName",
				Addr:     "defAddr",
				Employer: company{Name: "defCom"}},
		},
		{ // case 3, both args and file, arg should win
			def:   defCnf,
			fpath: defpath,
			args:  []string{"--name", "nameFromArg", "--jointime", "2016-12-02 12:03:04"},
			result: testStruct{
				JoinTime: time.Date(2016, 12, 2, 12, 3, 4, 0, time.UTC),
				Name:     "nameFromArg",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"}},

			// JoinTime: time.Date(2016, 12, 2, 12, 3, 4, 0, time.UTC),
		},
		{ // case 4, mix arg and default, arg should win
			def:   defCnf,
			fpath: "",
			args:  []string{"--name", "nameFromArg", "--employer-name", "argCom"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromArg",
				Addr:     "defAddr",
				Employer: company{Name: "argCom"}},
		},
		{ // case 5, specify nonexist config file, result should be default
			def:   defCnf,
			fpath: defpath,
			args:  []string{"--" + DefCfgFileFlagName, "dosntexist"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "defName",
				Addr:     "defAddr",
				Employer: company{Name: "defCom"}},
		},
		{ // case 6, specify nonexist config file and args, args should win
			def:   defCnf,
			fpath: defpath,
			args:  []string{"--" + DefCfgFileFlagName, "dosntexist", "--addr", "addrFromArg"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "defName",
				Addr:     "addrFromArg",
				Employer: company{Name: "defCom"}},
		},
		{ // case 7, use the default congi file, result should be value from file
			def:   defCnf,
			fpath: defpath,
			args:  []string{},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromFile",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"}},
		},
		{ // case 8, action test, action arg from cli, rest from file
			def:   defCnf,
			fpath: defpath,
			args:  []string{"act", "--netname", "disk1"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromFile",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"},
				Act:      struct{ NetName string }{NetName: "disk1"},
			},
		},
		{ // case 9, negative case, no cli, all from file
			def:        defCnf,
			expectFail: true,
			dontDoFlag: true,
			fpath:      defpath,
			args:       []string{"act", "--netname", "disk1"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromFile",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"},
				Act:      struct{ NetName string }{NetName: "disk1"},
			},
		},

		{ // case 9, no cli, all from file
			def:        defCnf,
			dontDoFlag: true,
			fpath:      defpath,
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromFile",
				Addr:     "addrFromFile",
				Employer: company{Name: "comFromFile"},
				Act:      struct{ NetName string }{NetName: ""},
			},
		},
	}
	for i, c := range caseList {
		t.Logf("testing case %d", i)
		err := doTest(t, c)
		if err != nil {
			t.Logf("case %d fails with err %v", i, err)
			if !c.expectFail {
				t.Fatal()
			} else {
				t.Logf("case %d failed as expected, %v", i, err)
			}
		} else {
			if !c.expectFail {
				t.Logf("case %d finished successfully", i)
			} else {
				t.Fatalf("case %d succeed while expect to fail", i)
			}
		}

	}

}

func discardOutput[X any](cnf *SConf[X]) {
	cnf.Filler.SetOut(&bytes.Buffer{})
	cnf.Filler.SetErr(&bytes.Buffer{})
}

func writeYAML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNilDefReturnsError(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("NewSConf panicked on a nil config pointer: %v", rec)
		}
	}()
	var def *struct{ Name string }
	if _, err := NewSConf(def, "app", "usage"); err == nil {
		t.Fatal("expected an error when def is nil")
	}
}

func TestYAMLReplacesSlicesAndMaps(t *testing.T) {
	type cfg struct {
		NumList []int
		Tags    map[string]string
	}
	path := writeYAML(t, "numlist: [9]\ntags:\n  a: file\n")
	def := cfg{
		NumList: []int{1, 2, 3},
		Tags:    map[string]string{"a": "default", "b": "stale"},
	}
	cnf, err := NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[cfg](path))
	if err != nil {
		t.Fatal(err)
	}
	discardOutput(cnf)
	if _, ferr, aerr := cnf.Read(nil); ferr != nil || aerr != nil {
		t.Fatalf("ferr=%v aerr=%v", ferr, aerr)
	}
	got := cnf.GetConf()
	if len(got.NumList) != 1 || got.NumList[0] != 9 {
		t.Fatalf("NumList = %#v, want [9]", got.NumList)
	}
	if len(got.Tags) != 1 || got.Tags["a"] != "file" {
		t.Fatalf("Tags = %#v, want map[a:file]", got.Tags)
	}

	emptyPath := writeYAML(t, "numlist: []\n")
	def = cfg{NumList: []int{1, 2, 3}}
	cnf, err = NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[cfg](emptyPath))
	if err != nil {
		t.Fatal(err)
	}
	discardOutput(cnf)
	if _, ferr, aerr := cnf.Read(nil); ferr != nil || aerr != nil {
		t.Fatalf("empty slice ferr=%v aerr=%v", ferr, aerr)
	}
	if got := cnf.GetConf().NumList; len(got) != 0 {
		t.Fatalf("empty YAML list left NumList = %#v", got)
	}
}

func TestDisabledFlagsDoNotPanic(t *testing.T) {
	type cfg struct {
		Name string
		Bad  []struct{ N int }
	}
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Read panicked with flag filling disabled: %v", rec)
		}
	}()
	path := writeYAML(t, "name: fromfile\n")
	def := cfg{Name: "def"}
	cnf, err := NewSConf(&def, "app", "usage",
		WithDefaultConfigFilePath[cfg](path),
		WithFillFlags[cfg](false),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, ferr, aerr := cnf.Read(nil)
	if ferr != nil || aerr != nil {
		t.Fatalf("ferr=%v aerr=%v", ferr, aerr)
	}
	if cnf.GetConf().Name != "fromfile" {
		t.Fatalf("Name = %q, want fromfile", cnf.GetConf().Name)
	}
}

func TestDisabledFlagsDoNotMutateSharedPointers(t *testing.T) {
	type cfg struct {
		Name string
		Note *string
	}
	path := writeYAML(t, "name: fromfile\n")
	note := "default"
	def := cfg{Name: "def", Note: &note}
	cnf, err := NewSConf(&def, "app", "usage",
		WithDefaultConfigFilePath[cfg](path),
		WithFillFlags[cfg](false),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, ferr, _ := cnf.Read([]string{"--note", "sneaky"})
	if ferr != nil {
		t.Fatalf("file error: %v", ferr)
	}
	if cnf.GetConf().Name != "fromfile" {
		t.Fatalf("Name = %q, want fromfile", cnf.GetConf().Name)
	}
	if got := *cnf.GetConf().Note; got != "default" {
		t.Fatalf("Note = %q, want default; flag filling is disabled", got)
	}
}

func TestFileLoadsWhenArgParseFails(t *testing.T) {
	type cfg struct {
		Name string `required:""`
		Addr string
	}
	path := writeYAML(t, "name: fromfile\naddr: fromfile\n")
	def := cfg{Name: "def", Addr: "def"}
	cnf, err := NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[cfg](path))
	if err != nil {
		t.Fatal(err)
	}
	discardOutput(cnf)
	_, ferr, aerr := cnf.Read(nil)
	if ferr != nil {
		t.Fatalf("file error: %v", ferr)
	}
	if aerr == nil {
		t.Fatal("expected a missing-required-flag error")
	}
	if cnf.GetConf().Addr != "fromfile" || cnf.GetConf().Name != "fromfile" {
		t.Fatalf("file was skipped: %+v", cnf.GetConf())
	}
}

func TestFileLoadsWhenFlagValueMissing(t *testing.T) {
	type plain struct {
		Name string
		Addr string
	}
	path := writeYAML(t, "name: fromfile\naddr: fromfile\n")
	def := plain{Name: "def", Addr: "def"}
	cnf, err := NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[plain](path))
	if err != nil {
		t.Fatal(err)
	}
	discardOutput(cnf)
	_, ferr, aerr := cnf.Read([]string{"--name"}) // value missing
	if ferr != nil {
		t.Fatalf("file error: %v", ferr)
	}
	if aerr == nil {
		t.Fatal("expected an argument error")
	}
	if cnf.GetConf().Addr != "fromfile" {
		t.Fatalf("Addr = %q, want fromfile", cnf.GetConf().Addr)
	}
}

func TestSecondReadDropsStaleFileValues(t *testing.T) {
	type plain struct {
		Name string
		Addr string
	}
	dir := t.TempDir()
	first := filepath.Join(dir, "a.yaml")
	second := filepath.Join(dir, "b.yaml")
	if err := os.WriteFile(first, []byte("name: A\naddr: A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("name: B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	def := plain{Name: "def", Addr: "def"}
	cnf, err := NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[plain](first))
	if err != nil {
		t.Fatal(err)
	}
	discardOutput(cnf)
	if _, ferr, aerr := cnf.Read(nil); ferr != nil || aerr != nil {
		t.Fatalf("first read ferr=%v aerr=%v", ferr, aerr)
	}
	if _, ferr, aerr := cnf.Read([]string{"--" + DefCfgFileFlagName, second}); ferr != nil || aerr != nil {
		t.Fatalf("second read ferr=%v aerr=%v", ferr, aerr)
	}
	got := cnf.GetConf()
	if got.Name != "B" || got.Addr != "def" {
		t.Fatalf("after second file, Name=%q Addr=%q, want Name=B Addr=def", got.Name, got.Addr)
	}
}

func TestCLIOverridesPointerFieldsAfterYAML(t *testing.T) {
	type inner struct {
		Name string
	}
	type cfg struct {
		Addr     *string
		Employer *inner
	}
	path := writeYAML(t, "addr: fromfile\nemployer:\n  name: filecom\n")
	addr := "defaddr"
	employer := &inner{Name: "defcom"}
	def := cfg{Addr: &addr, Employer: employer}
	cnf, err := NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[cfg](path))
	if err != nil {
		t.Fatal(err)
	}
	discardOutput(cnf)
	_, ferr, aerr := cnf.Read([]string{"--addr", "fromcli", "--employer-name", "clicom"})
	if ferr != nil || aerr != nil {
		t.Fatalf("ferr=%v aerr=%v", ferr, aerr)
	}
	got := cnf.GetConf()
	if got.Addr != &addr || *got.Addr != "fromcli" {
		t.Fatalf("Addr = %q, pointer preserved = %v", *got.Addr, got.Addr == &addr)
	}
	if got.Employer != employer || got.Employer.Name != "clicom" {
		t.Fatalf("Employer = %+v, pointer preserved = %v", got.Employer, got.Employer == employer)
	}
}

func TestConfigFlagNameCollisionReturnsError(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("NewSConf panicked on a flag name collision: %v", rec)
		}
	}()
	def := struct {
		CfgFromFile string
	}{}
	_, err := NewSConf(&def, "app", "usage", WithDefaultConfigFilePath[struct{ CfgFromFile string }]("cfg.yaml"))
	if err == nil {
		t.Fatal("expected an error when a field flag is also named cfgfromfile")
	}
}
