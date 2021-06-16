package handlers

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gorm.io/gorm"
	"strconv"
)

type InfoSceneResp struct {
	orm.Scene
	CreateTime      int64           `json:"create_time"`
	EffectStartTime int64           `json:"effect_start_time"`
	EffectEndTime   int64           `json:"effect_end_time"`
	SceneConditions []ConditionInfo `json:"scene_conditions"`
	SceneTasks      []SceneTaskInfo `json:"scene_tasks"`
}

type ConditionInfo struct {
	orm.ConditionInfo
	DeviceInfo `json:"device_info"`
}

type SceneTaskInfo struct {
	orm.SceneTask
	ControlSceneInfo ControlSceneInfo `json:"control_scene_info"`
	DeviceInfo       `json:"device_info"`
}

type ControlSceneInfo struct {
	Name   string      `json:"name"`
	Status sceneStatus `json:"status"`
}

type DeviceInfo struct {
	Name         string       `json:"name"`
	LocationName string       `json:"location_name"`
	LogoURL      string       `json:"logo_url"`
	Status       deviceStatus `json:"status"`
}

func InfoScene(c *gin.Context) {
	var (
		err     error
		resp    InfoSceneResp
		sceneID int
		scene   orm.Scene
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if sceneID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	scene, err = orm.GetSceneInfoById(sceneID)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.SceneNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if err = WrapResponse(scene, &resp); err != nil {
		return
	}

	return
}

func WrapResponse(scene orm.Scene, resp *InfoSceneResp) (err error) {

	var (
		conditionInfo   ConditionInfo
		sceneTask       SceneTaskInfo
		sceneConditions = make([]ConditionInfo, 0)
		sceneTasks      = make([]SceneTaskInfo, 0)
	)

	for _, condition := range scene.SceneConditions {

		if conditionInfo, err = WrapConditionInfo(condition); err != nil {
			return
		}

		sceneConditions = append(sceneConditions, conditionInfo)
	}

	for _, task := range scene.SceneTasks {
		if sceneTask, err = WrapTaskInfo(task); err != nil {
			return
		}
		sceneTasks = append(sceneTasks, sceneTask)
	}

	resp.SceneConditions = sceneConditions
	resp.Scene = scene
	resp.SceneTasks = sceneTasks
	resp.CreateTime = scene.CreatedAt.Unix()
	resp.EffectStartTime = scene.EffectStart.Unix()
	resp.EffectEndTime = scene.EffectEnd.Unix()

	return
}

func WrapConditionInfo(condition orm.SceneCondition) (conditionInfo ConditionInfo, err error) {
	var (
		deviceInfo DeviceInfo
	)

	if condition.ConditionType == orm.ConditionTypeDeviceStatus {
		if deviceInfo, err = WrapDeviceInfo(condition.DeviceID); err != nil {
			return
		}
	}

	conditionInfo.Timing = condition.TimingAt.Unix()
	conditionInfo.SceneCondition = condition
	conditionInfo.DeviceInfo = deviceInfo

	return
}

func WrapTaskInfo(task orm.SceneTask) (taskInfo SceneTaskInfo, err error) {
	var (
		scene orm.Scene
	)
	taskInfo = SceneTaskInfo{
		SceneTask: task,
	}

	if task.Type != orm.TaskTypeSmartDevice {
		if scene, err = orm.GetDeletedSceneByID(task.ControlSceneID); err != nil {
			if errors2.Is(err, gorm.ErrRecordNotFound) {
				err = errors.Wrap(err, errors.SceneNotExist)
				return
			}
			err = errors.Wrap(err, errors.InternalServerErr)
			return
		}
		taskInfo.ControlSceneInfo.Name = scene.Name
		// 场景已被删除
		if scene.Deleted.Valid {
			taskInfo.ControlSceneInfo.Status = sceneAlreadyDelete
			return
		}
		taskInfo.ControlSceneInfo.Status = sceneNormal
		return
	}

	if taskInfo.DeviceInfo, err = WrapDeviceInfo(taskInfo.SceneTaskDevices[0].DeviceID); err != nil {
		return
	}

	return

}

func WrapDeviceInfo(deviceID int) (deviceInfo DeviceInfo, err error) {
	var (
		location orm.Location
		device   orm.Device
	)

	if device, err = orm.GetDeletedDeviceByID(deviceID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.DeviceNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if device.LocationID != 0 {
		if location, err = orm.GetLocationByID(device.LocationID); err != nil {
			return
		}
	}
	deviceInfo.Name = device.Name
	di, _ := plugin.SupportedDeviceInfo[device.Model]
	deviceInfo.LogoURL = di.LogoURL
	deviceInfo.LocationName = location.Name

	if device.Deleted.Valid {
		// 设备已删除
		deviceInfo.Status = deviceAlreadyDelete
		return
	}

	// TODO:判断设备是否可连接
	deviceInfo.Status = deviceNormal
	return

}
