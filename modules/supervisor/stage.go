package supervisor

import (
	"io/ioutil"
	"os"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
)

type StageVal string

const (
	stageFile = "stage.json"
)

var (
	StageBackupInit  = StageVal("back_init")
	StageRestoreInit = StageVal("restore_init")
)

// Stage 阶段过程描述
type Stage struct {
	dir        string
	Value      StageVal `json:"value"`
	BackupName string   `json:"backup_name"`
}

func NewStage(dir string, val StageVal) *Stage {
	return &Stage{
		dir:   dir,
		Value: val,
	}
}

func loadStage(dir string) (stage *Stage, err error) {
	fn := filepath.Join(dir, stageFile)
	fd, err := os.Open(fn)
	if err != nil {
		return
	}
	defer fd.Close()
	content, err := ioutil.ReadAll(fd)
	if err != nil {
		return
	}
	stage = &Stage{}
	err = jsoniter.Unmarshal(content, stage)
	stage.dir = dir
	return
}

func (s Stage) save() (err error) {
	_ = os.MkdirAll(s.dir, os.ModePerm)
	content, err := jsoniter.MarshalIndent(s, "", "    ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filepath.Join(s.dir, stageFile), content, 0666)
	return
}

func (s Stage) remove() error {
	return os.RemoveAll(s.dir)
}
