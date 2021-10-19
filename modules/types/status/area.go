package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与家庭相关的响应状态码
const (
	AreaNotExist = iota + 1000
	OwnerQuitErr
	AreaNameInputNilErr
	AreaNameLengthLimit
	SABindError
)

func init() {
	errors.NewCode(AreaNotExist, "该家庭不存在")
	errors.NewCode(OwnerQuitErr, "当前家庭创建者不允许退出家庭")
	errors.NewCode(AreaNameInputNilErr, "请输入家庭名称")
	errors.NewCode(AreaNameLengthLimit, "家庭名称长度不能超过30")
	errors.NewCode(SABindError, "SA绑定失败")
}
