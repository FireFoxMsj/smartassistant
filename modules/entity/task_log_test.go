package entity

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestNewTaskLog(t *testing.T) {
	scene := Scene{
		Name:    "test1",
		AutoRun: true,
	}
	parentID := "1"
	err := NewTaskLog(scene, parentID, nil)
	assert.NoError(t, err)

	device := Device{
		Name:       "test11",
		LocationID: 666,
	}
	location := Location{
		ID:   666,
		Name: "客厅",
	}
	GetDB().Create(&location)
	err = NewTaskLog(device, "2", &parentID)
	assert.NoError(t, err)
}

func TestUpdateTaskLog(t *testing.T) {
	err := UpdateTaskLog("1", errors.New("undefinedError"))
	assert.NoError(t, err)
	var taskLog TaskLog
	GetDB().Where("id=?", "1").Find(&taskLog)
	assert.Equal(t, taskLog.Result, TaskFail)

	err = UpdateTaskLog("1", nil)
	assert.NoError(t, err)
	GetDB().Where("id=?", "1").Find(&taskLog)
	assert.Equal(t, taskLog.Result, TaskSuccess)
}

func TestUpdateParentLog(t *testing.T) {
	err := UpdateParentLog("1")
	assert.NoError(t, err)

	err = UpdateTaskLog("2", errors.New("undefinedError"))
	assert.NoError(t, err)
	var taskLog TaskLog
	GetDB().Where("id=?", "1").Find(&taskLog)
	assert.Equal(t, taskLog.Result, TaskFail)
}
