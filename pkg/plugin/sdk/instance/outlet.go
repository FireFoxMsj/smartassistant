package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type Outlet struct {
	Power *attribute.Power `tag:"name:power;required"`
	// InUse attribute.InUse
	Name *attribute.Name
}

func (o Outlet) InstanceName() string {
	return "outlet"
}

func NewOutlet() Outlet {
	return Outlet{
		Power: attribute.NewPower(),
		Name:  &attribute.Name{},
	}
}
