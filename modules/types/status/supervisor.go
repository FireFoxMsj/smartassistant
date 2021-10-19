package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与系统管理相关的响应状态码

const (
	FileNotExistErr = iota + 7000
)

func init() {
	errors.NewCode(FileNotExistErr, "文件不存在")

}
