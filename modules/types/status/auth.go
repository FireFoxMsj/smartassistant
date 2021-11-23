package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与授权相关的状态码

const (
	ErrInvalidGrantType = iota + 8000
	ErrInvalidRefreshToken
)

func init() {
	errors.NewCode(ErrInvalidGrantType, "无效的授权类型")
	errors.NewCode(ErrInvalidRefreshToken, "无效的refresh token")
}
