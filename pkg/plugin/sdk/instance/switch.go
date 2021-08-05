package instance

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
)

type Switch struct {
	Power *attribute.Power `tag:"name:power;required"`
	Name  *attribute.Name
}

func (l Switch) InstanceName() string {
	return "switch"
}

func NewSwitch() Switch {
	return Switch{
		Power: attribute.NewPower(),
		Name:  &attribute.Name{},
	}
}

type Device struct {
	Light1  LightBulb
	Switch1 Switch
}
