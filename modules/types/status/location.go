package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与房间相关的响应状态码
const (
	LocationNotExit = iota + 3000
	LocationNameInputNilErr
	LocationNameLengthLimit
	LocationNameExist
)

func init() {
	errors.NewCode(LocationNotExit, "该房间不存在")
	errors.NewCode(LocationNameInputNilErr, "请输入房间名称")
	errors.NewCode(LocationNameLengthLimit, "房间名称长度不能超过20")
	errors.NewCode(LocationNameExist, "房间名称重复")
}
