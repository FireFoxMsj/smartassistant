package handlers

import (
	"sort"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

const logSizeDefault = 40

type ListSceneTaskReq struct {
	Start int `form:"start"`
	Size  int `form:"size"`
}

type ListSceneTaskLogResp []DateLogInfo

type DateLogInfo struct {
	Date              string             `json:"date"`
	SceneTaskLogInfos []SceneTaskLogInfo `json:"items"`
}

type SceneTaskLogInfo struct {
	Name       string             `json:"name"`
	Type       orm.TaskType       `json:"type"`
	Result     orm.TaskResultType `json:"result"`
	FinishedAt int64              `json:"finished_at"`
	Items      []TaskLogItem      `json:"items"`
}

type TaskLogItem struct {
	Name         string             `json:"name"`
	Type         orm.TaskType       `json:"type"`
	LocationName string             `json:"location_name"`
	Result       orm.TaskResultType `json:"result"`
}

func ListSceneTaskLog(c *gin.Context) {
	var (
		err error
		req ListSceneTaskReq

		taskLogs []orm.TaskLog
		resp     ListSceneTaskLogResp
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if req.Size == 0 {
		req.Size = logSizeDefault
	}

	if err = orm.GetDB().
		Preload("ChildTaskLogs").
		Order("finished_at desc").
		Where("type !=? and finish=1 and result !=?",
			orm.TaskTypeSmartDevice, orm.TaskSceneAlreadyDeleted).
		Offset(req.Start).
		Limit(req.Size).
		Find(&taskLogs).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if resp, err = LogInfosGroupByDate(taskLogs); err != nil {
		return
	}

	return
}

func LogInfosGroupByDate(taskLogs []orm.TaskLog) (logInfos []DateLogInfo, err error) {
	var (
		logDates     []string
		dateLogInfos = make(map[string][]SceneTaskLogInfo)
	)

	logInfos = make([]DateLogInfo, 0)

	for _, taskLog := range taskLogs {
		if !taskLog.Finish {
			continue
		}

		taskLogInfo := SceneTaskLogInfo{
			Name:       taskLog.Name,
			Type:       taskLog.Type,
			Result:     taskLog.Result,
			FinishedAt: taskLog.FinishedAt.Unix(),
			Items:      WrapLogItems(taskLog),
		}
		date := taskLog.FinishedAt.Format("2006-01")
		if _, ok := dateLogInfos[date]; !ok {
			dateLogInfos[date] = make([]SceneTaskLogInfo, 0)
		}
		dateLogInfos[date] = append(dateLogInfos[date], taskLogInfo)
	}

	for d, _ := range dateLogInfos {
		logDates = append(logDates, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(logDates)))
	for _, d := range logDates {
		logInfos = append(logInfos, DateLogInfo{
			Date:              d,
			SceneTaskLogInfos: dateLogInfos[d],
		})
	}
	return
}

func WrapLogItems(taskLog orm.TaskLog) (taskItems []TaskLogItem) {

	taskItems = make([]TaskLogItem, 0)

	// 任务部分执行成功/执行失败时展示执行详情
	if taskLog.Result == orm.TaskPartSuccess || taskLog.Result == orm.TaskFail && len(taskLog.ChildTaskLogs) != 0 {
		for _, taskLog := range taskLog.ChildTaskLogs {
			taskItems = append(taskItems, TaskLogItem{
				Name:         taskLog.Name,
				Type:         taskLog.Type,
				Result:       taskLog.Result,
				LocationName: taskLog.DeviceLocation,
			})
		}
	}

	return
}
