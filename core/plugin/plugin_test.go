package plugin_test

import (
	"fmt"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"testing"
)

func TestInstall(t *testing.T) {
	plg := plugin.Plugin{}
	plg.DownloadURL = "http://tysq2.yctc.tech/api/file/originals/id/2009037/fn/plugin1.zip"
	plg.Name = "plugin3"
	err := plugin.Install(plg)
	if err != nil {
		t.FailNow()
	}
}

func TestList(t *testing.T) {
	plgs, err := plugin.List()
	if err != nil {
		t.FailNow()
	}

	for _, plg := range plgs {
		fmt.Println(plg.Name)
	}
}

func TestInfo(t *testing.T) {
	var ID = "001"
	plg, err := plugin.Info(ID)
	if err != nil {
		t.FailNow()
	}
	if plg.ID != ID {
		t.FailNow()
	}

}
