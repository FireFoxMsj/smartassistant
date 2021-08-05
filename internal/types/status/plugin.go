package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与插件相关的响应状态码
const (
	PluginDomainNotExist = iota + 6000
	PluginServiceNotExist
)

func init() {
	errors.NewCode(PluginDomainNotExist, "插件不存在")
	errors.NewCode(PluginServiceNotExist, "插件功能不存在")
}
