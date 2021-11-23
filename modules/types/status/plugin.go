package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与插件相关的响应状态码
const (
	PluginDomainNotExist = iota + 6000
	PluginServiceNotExist
	PluginTypeNotSupport
	PluginIsEmpty
	PluginContentIllegal
)

func init() {
	errors.NewCode(PluginDomainNotExist, "插件不存在")
	errors.NewCode(PluginServiceNotExist, "插件功能不存在")
	errors.NewCode(PluginTypeNotSupport, "插件包格式不正确")
	errors.NewCode(PluginIsEmpty, "请上传插件")
	errors.NewCode(PluginContentIllegal, "插件包内容不符合规范")
}
