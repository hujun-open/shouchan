package shouchan

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/hujun-open/myflags"
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
	} `action:""`
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

	options := []SconfOption[*testStruct]{}
	options = append(options, WithDefaultConfigFilePath[*testStruct](setup.fpath))
	options = append(options, WithFillFlags[*testStruct](!setup.dontDoFlag))
	if setup.expectFail {
		options = append(options, WithFillOptions[*testStruct]([]myflags.FillerOption{
			myflags.WithFlagErrHandling(flag.ContinueOnError),
		}))
	}
	cnf, err := NewSConf(&setup.def, "test", "golangdevtest", options...)
	if err != nil {
		return err
	}
	ferr, aerr := cnf.Read(setup.args)
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
			args:  []string{DefCfgFileFlagName, defpath},
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
			args:  []string{"-name", "nameFromArg", "-jointime", "2016-12-02 12:03:04"},
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
			args:  []string{"-name", "nameFromArg", "-employer-name", "argCom"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "nameFromArg",
				Addr:     "defAddr",
				Employer: company{Name: "argCom"}},
		},
		{ // case 5, specify nonexist config file, result should be default
			def:   defCnf,
			fpath: defpath,
			args:  []string{DefCfgFileFlagName, "dosntexist"},
			result: testStruct{
				JoinTime: time.Date(1999, 1, 2, 3, 4, 5, 0, time.UTC),
				Name:     "defName",
				Addr:     "defAddr",
				Employer: company{Name: "defCom"}},
		},
		{ // case 6, specify nonexist config file and args, args should win
			def:   defCnf,
			fpath: defpath,
			args:  []string{DefCfgFileFlagName, "dosntexist", "-addr", "addrFromArg"},
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
			args:  []string{"act", "-netname", "disk1"},
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
			args:       []string{"act", "-netname", "disk1"},
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
