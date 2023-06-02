package shouchan

import (
	"fmt"
	"reflect"
	"testing"
)

type testActionSetup struct {
	args         []string
	expectResult struct {
		action string
		cfg    any
	}
	shouldFail bool
}

type zipShowCfg struct {
	Zipf, Fname string
}

func (zscfg *zipShowCfg) Default() *zipShowCfg {
	return &zipShowCfg{
		Zipf:  "defaultShowZipf",
		Fname: "defaultShowFname",
	}
}

func (zacfg zipShowCfg) DefaultCfgFilePath() string {
	return ""
}

type zipArchiveCfg struct {
	Folder, Zipf string
}

func (zacfg *zipArchiveCfg) Default() *zipArchiveCfg {
	return &zipArchiveCfg{
		Zipf:   "defaultArchiveZipf",
		Folder: "defaultArchiveFolder",
	}
}

func (zacfg zipArchiveCfg) DefaultCfgFilePath() string {
	return ""
}

func doActionTest(t *testing.T, setup testActionSetup) error {
	acnf, err := NewActionConfWithCMDLine(map[string]ConfigWithDefCfgFilePath{
		"show": (&zipShowCfg{}).Default(),
		"zip":  (&zipArchiveCfg{}).Default(),
	})
	if err != nil {
		return err
	}
	err, _, _ = acnf.Read(setup.args)
	if err != nil {
		return err
	}
	if acnf.GetLoadedAction() != setup.expectResult.action {
		return fmt.Errorf("expected action is %v, but get %v", setup.expectResult.action, acnf.GetLoadedAction())
	}
	if !reflect.DeepEqual(acnf.GetLoadedConf(), setup.expectResult.cfg) {
		return fmt.Errorf("expect cfg is %+v, but get %+v", setup.expectResult.cfg, acnf.GetLoadedConf())
	}
	return nil
}

func TestAction(t *testing.T) {
	caseList := []testActionSetup{
		//case 0, default
		{
			args: []string{"zip"},
			expectResult: struct {
				action string
				cfg    any
			}{
				action: "zip",
				cfg: &zipArchiveCfg{
					Zipf:   "defaultArchiveZipf",
					Folder: "defaultArchiveFolder",
				},
			},
		},
		//case 1, wrong action
		{
			args: []string{"wrong"},
			expectResult: struct {
				action string
				cfg    any
			}{
				action: "zip",
				cfg: &zipArchiveCfg{
					Zipf:   "defaultArchiveZipf",
					Folder: "defaultArchiveFolder",
				},
			},
			shouldFail: true,
		},
		//case 2, arg
		{
			args: []string{"zip", "-zipf", "newzipf"},
			expectResult: struct {
				action string
				cfg    any
			}{
				action: "zip",
				cfg: &zipArchiveCfg{
					Zipf:   "newzipf",
					Folder: "defaultArchiveFolder",
				},
			},
		},
		//case 3, arg
		{
			args: []string{"zip", "-zipf", "newzipf", "-folder", "newfolder"},
			expectResult: struct {
				action string
				cfg    any
			}{
				action: "zip",
				cfg: &zipArchiveCfg{
					Zipf:   "newzipf",
					Folder: "newfolder",
				},
			},
		},
		//case 3, using show
		{
			args: []string{"show", "-zipf", "newzipf", "-fname", "newzipinside"},
			expectResult: struct {
				action string
				cfg    any
			}{
				action: "show",
				cfg: &zipShowCfg{
					Zipf:  "newzipf",
					Fname: "newzipinside",
				},
			},
		},
	}

	for i, c := range caseList {
		err := doActionTest(t, c)
		if err != nil {
			if !c.shouldFail {
				t.Fatalf("case %d failed, %v", i, err)
			} else {
				t.Logf("case %d failed as expected, %v", i, err)
			}
		} else {
			t.Logf("case %d finished successfully", i)
		}
	}
}
