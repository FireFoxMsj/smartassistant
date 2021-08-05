package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type LightBulb struct {
	Power      *attribute.Power `tag:"name:power;required"`
	ColorTemp  *attribute.ColorTemp
	Brightness *attribute.Brightness
	Name       *attribute.Name
}

func (l LightBulb) InstanceName() string {
	return "light_bulb"
}

func NewLightBulb() LightBulb {
	return LightBulb{
		Power:      attribute.NewPower(),
		ColorTemp:  &attribute.ColorTemp{},
		Brightness: &attribute.Brightness{},
		Name:       &attribute.Name{},
	}
}
