package handlers

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"gorm.io/gorm"
)

type sceneStatus int

const (
	sceneNormal sceneStatus = iota + 1
	sceneAlreadyDelete
)

type deviceStatus int

const (
	deviceNormal deviceStatus = iota + 1
	deviceAlreadyDelete
	deviceDisConnect
)

type listType int

const (
	allScene listType = iota
	permitScene
)

type sceneListReq struct {
	Type listType `form:"type"`
}

type sceneListResp struct {
	Manual  []manualSceneInfo  `json:"manual"`
	AutoRun []autoRunSceneInfo `json:"auto_run"`
}

type Scene struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	ControlPermission bool   `json:"control_permission"`
}

type manualSceneInfo struct {
	Scene
	Items []Item `json:"items"`
}

type autoRunSceneInfo struct {
	Scene
	IsOn      bool           `json:"is_on"`
	Condition sceneCondition `json:"condition"`
	Items     []Item         `json:"items"`
}

type sceneCondition struct {
	Type    orm.ConditionType `json:"type"`
	LogoURL string            `json:"logo_url"`
	Status  int               `json:"status"`
}

type Item struct {
	ID      int          `json:"-"`
	Type    orm.TaskType `json:"type"`
	LogoURL string       `json:"logo_url"`
	Status  int          `json:"status"`
	devices []orm.SceneTaskDevice
}

func ListScene(c *gin.Context) {
	var (
		err    error
		req    sceneListReq
		resp   sceneListResp
		scenes []orm.Scene
		user   *session.User
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	user = session.Get(c)

	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if scenes, err = orm.GetScenes(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if resp.Manual, resp.AutoRun, err = WrapScenes(scenes, user.UserID, req.Type); err != nil {
		return
	}

	return
}

func WrapScenes(scenes []orm.Scene, userID int, listType listType) (manualScenes []manualSceneInfo, autoRunScenes []autoRunSceneInfo, err error) {
	var (
		items             []Item
		condition         sceneCondition
		controlPermission bool
	)

	manualScenes = make([]manualSceneInfo, 0)
	autoRunScenes = make([]autoRunSceneInfo, 0)

	for _, scene := range scenes {

		if controlPermission, err = CheckControlPermission(scene.ID, userID); err != nil {
			return
		}
		if listType == permitScene && !controlPermission {
			continue
		}

		// 场景执行任务信息
		if items, err = WrapItems(scene.ID); err != nil {
			return
		}

		// 场景信息
		sceneInfo := Scene{
			ID:                scene.ID,
			Name:              scene.Name,
			ControlPermission: controlPermission,
		}

		// 场景触发条件信息()
		if scene.AutoRun {
			// 自动触发条件
			var canControl bool
			if condition, canControl, err = WrapCondition(scene.ID, userID); err != nil {
				return
			}
			// 没有触发条件中设备的控制权限，ControlPermission为false
			if !canControl {
				sceneInfo.ControlPermission = false
			}

			autoRunScene := autoRunSceneInfo{
				Scene:     sceneInfo,
				Items:     items,
				Condition: condition,
				IsOn:      scene.IsOn,
			}
			autoRunScene.Condition = condition
			autoRunScenes = append(autoRunScenes, autoRunScene)

		} else {
			// 手动没有触发条件
			manualScene := manualSceneInfo{
				Scene: sceneInfo,
				Items: items,
			}
			manualScenes = append(manualScenes, manualScene)
		}

	}
	return
}

func WrapCondition(sceneID, userID int) (sceneCondition sceneCondition, canControlDevice bool, err error) {
	var (
		conditions    []orm.SceneCondition
		conditionItem orm.ConditionItem
	)

	canControlDevice = true
	// 获取场景的所有触发条件
	if conditions, err = orm.GetConditionsBySceneID(sceneID); err != nil {
		return
	}

	for i, c := range conditions {
		// 只返回第一个触发条件的信息
		sceneCondition.Type = conditions[0].ConditionType
		if c.ConditionType == orm.ConditionTypeDeviceStatus {
			// 智能设备触发条件
			// 判断对应权限
			conditionItem, err = orm.GetConditionItemByConditionID(c.ID)
			if err != nil {
				canControlDevice = false
				return
			}
			if !orm.IsDeviceControlPermit(userID, c.DeviceID, conditionItem.Attribute) {
				canControlDevice = false
				return
			}

			// 第一个触发条件为设备时，包装对应信息
			if i == 0 {
				item := Item{ID: c.DeviceID}
				if err = WrapDeviceItem(&item); err != nil {
					return
				}
				sceneCondition.LogoURL = item.LogoURL
				sceneCondition.Status = item.Status
			}
		}
	}
	return
}

func WrapItems(sceneID int) (items []Item, err error) {
	var (
		tasks []orm.SceneTask
		item  Item
	)

	items = make([]Item, 0)
	// 获取场景所有执行任务
	if tasks, err = orm.GetSceneTasksBySceneID(sceneID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	for _, task := range tasks {
		if item, err = WrapItem(task); err != nil {
			return
		}
		items = append(items, item)

	}

	return
}

func WrapItem(task orm.SceneTask) (item Item, err error) {
	var (
		taskDevices []orm.SceneTaskDevice
		scene       orm.Scene
	)

	item.Type = task.Type

	// 执行任务类型为智能设备
	if task.Type == orm.TaskTypeSmartDevice {
		// 拿出这个执行任务中的所有设备操作
		if taskDevices, err = orm.GetSceneTaskItemsByTaskID(task.ID); err != nil {
			return
		}
		item.ID = taskDevices[0].DeviceID
		item.devices = taskDevices
		if err = WrapDeviceItem(&item); err != nil {
			return
		}
		return
	}
	// 执行任务类型为场景
	item.ID = task.ControlSceneID
	if scene, err = orm.GetDeletedSceneByID(task.ControlSceneID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.SceneNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if scene.Deleted.Valid {
		item.Status = int(sceneAlreadyDelete)
		return
	}
	item.Status = int(sceneNormal)
	return
}

func CheckControlPermission(sceneID int, userID int) (controlPermission bool, err error) {
	var (
		items []Item
	)

	// 没有控制场景的权限，直接返回
	if !orm.JudgePermit(userID, permission.SceneControl) {
		return
	}
	controlPermission = true

	if items, err = WrapItems(sceneID); err != nil {
		return
	}
	for _, item := range items {
		// 校验执行任务为智能设备时对该设备的控制权限
		if item.Type == orm.TaskTypeSmartDevice {

			// 已删除的设备跳过判断
			if item.Status == int(deviceAlreadyDelete) {
				continue
			}
			// 判断设备每一个操作的控制权限
			for _, device := range item.devices {
				if !orm.IsDeviceControlPermit(userID, item.ID, device.Attribute) {
					controlPermission = false
					return
				}
			}
			continue
		}

		if controlPermission, err = CheckControlPermission(item.ID, userID); err != nil {
			return
		}
		// 嵌套控制场景不满足权限就直接返回false
		if !controlPermission {
			return
		}
	}

	return
}

func WrapDeviceItem(item *Item) (err error) {
	var (
		device orm.Device
	)

	if device, err = orm.GetDeletedDeviceByID(item.ID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.DeviceNotExist)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	di, _ := plugin.SupportedDeviceInfo[device.Model]
	item.LogoURL = di.LogoURL

	if device.Deleted.Valid {
		// 设备已删除
		item.Status = int(deviceAlreadyDelete)
		return
	}
	// TODO:判断设备是否可连接
	item.Status = int(deviceNormal)
	return
}
