package instance

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
)

type Switch struct {
	Power *attribute.Power `tag:"required"`
}

func (l Switch) InstanceName() string {
	return "switch"
}
