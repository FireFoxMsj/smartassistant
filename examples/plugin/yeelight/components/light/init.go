package light

import (
	"gitlab.yctc.tech/root/smartassistent.git/core"
	"gitlab.yctc.tech/root/smartassistent.git/examples/plugin/yeelight/components"
)

type LightEntity interface {
	components.ToggleEntity
	SetBrightness(args core.M) error
	SetColorTemp(args core.M) error
}
