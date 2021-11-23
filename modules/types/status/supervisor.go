package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与系统管理相关的响应状态码

const (
	FileNotExistErr = iota + 7000
	FileIsDirErr
	ImagePullErr
	GetImageVersionErr
)

func init() {
	errors.NewCode(FileNotExistErr, "文件不存在")
	errors.NewCode(FileIsDirErr, "非法访问")
	errors.NewCode(ImagePullErr, "拉取镜像失败")
	errors.NewCode(GetImageVersionErr, "获取版本信息失败")

}
