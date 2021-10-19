package task

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func TestMain(m *testing.M) {
	config.TestSetup()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// 启动一个任务，防止 queue 启动即休眠
	sFunc := func(task *Task) error {
		logger.Println("running startup task after 59 sec")
		return nil
	}
	ta := NewTask(sFunc, 59*time.Second)
	GetManager().(*LocalManager).pushTask(ta, "task to keep queue running")

	go GetManager().(*LocalManager).queue.start(ctx)
	code := m.Run()
	<-ctx.Done()
	config.TestTeardown()
	os.Exit(code)
}

// 测试Task运行，并且能正确记录日志
func TestNewTask(t *testing.T) {
	tFuncRun := false
	tFunc := func(task *Task) error {
		tFuncRun = true
		return nil
	}
	ta := NewTask(tFunc, 2*time.Second)
	GetManager().(*LocalManager).pushTask(ta, "test new task")
	time.Sleep(3 * time.Second)
	assert.True(t, tFuncRun, "test new task not run")
	var taskLogs []entity.TaskLog
	err := entity.GetDB().Where("task_id=?", ta.ID).Find(&taskLogs).Error
	assert.Nil(t, err)
	assert.NotEmpty(t, len(taskLogs), "task log not found")
}

func TestNewTaskSameDelay(t *testing.T) {
	tFuncRunA := false
	tFuncRunB := false
	tFuncA := func(task *Task) error {
		tFuncRunA = true
		return nil
	}
	tFuncB := func(task *Task) error {
		tFuncRunB = true
		return nil
	}
	ta := NewTask(tFuncA, 2*time.Second)
	tb := NewTask(tFuncB, 2*time.Second)
	GetManager().(*LocalManager).pushTask(ta, "test new task a")
	GetManager().(*LocalManager).pushTask(tb, "test new task b")
	time.Sleep(3 * time.Second)
	assert.True(t, tFuncRunA, "test new task a not run")
	assert.True(t, tFuncRunB, "test new task b not run")
}

func TestNewTaskAt(t *testing.T) {
	tFuncRun := false
	tFunc := func(task *Task) error {
		tFuncRun = true
		return nil
	}
	ta := NewTaskAt(tFunc, time.Now().Add(2*time.Second))
	GetManager().(*LocalManager).pushTask(ta, "test new task at")
	time.Sleep(3 * time.Second)
	assert.True(t, tFuncRun, "test new task at run")
}

func TestTaskWithParent(t *testing.T) {
	tFuncPRun := false
	tFuncCRun := false
	// 子task
	tFuncC := func(task *Task) error {
		tFuncCRun = true
		return nil
	}
	// 父task
	tFuncP := func(task *Task) error {
		tFuncPRun = true
		tc := NewTask(tFuncC, 2*time.Second).WithParent(task)
		GetManager().(*LocalManager).pushTask(tc, "test child task")
		return nil
	}
	tp := NewTask(tFuncP, 1*time.Second)
	GetManager().(*LocalManager).pushTask(tp, "test parent task")
	time.Sleep(2 * time.Second)
	assert.True(t, tFuncPRun)
	assert.False(t, tFuncCRun)
	time.Sleep(2 * time.Second)
	assert.True(t, tFuncCRun)
	var taskLogs []entity.TaskLog
	err := entity.GetDB().Where("parent_task_id=?", tp.ID).Find(&taskLogs).Error
	assert.Nil(t, err)
	assert.NotEmpty(t, len(taskLogs), "task log not found")
}

func TestTaskWithWrapper(t *testing.T) {
	tFuncRun := false
	tFuncWrapRun := false
	wFunc := func() WrapperFunc {
		return func(taskFunc TaskFunc) TaskFunc {
			tFuncWrapRun = true
			return func(task *Task) error {
				t.Logf("task index %v", task.index)
				return taskFunc(task)
			}
		}
	}
	tFunc := func(task *Task) error {
		tFuncRun = true
		return nil
	}
	ta := NewTask(tFunc, 1*time.Second).WithWrapper(wFunc())
	GetManager().(*LocalManager).pushTask(ta, "test task with wrapper")
	time.Sleep(3 * time.Second)
	assert.True(t, tFuncRun, "task not run")
	assert.True(t, tFuncWrapRun, "wrapper not run")
}
