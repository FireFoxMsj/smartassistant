package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type Outlet struct {
	Power *attribute.Power `tag:"required"`
	// InUse attribute.InUse
}

func (o Outlet) InstanceName() string {
	return "outlet"
}
