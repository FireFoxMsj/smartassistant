package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type Info struct {
	Name         *attribute.Name
	Identity     *attribute.Identity     `tag:"required"`
	Model        *attribute.Model        `tag:"required"`
	Manufacturer *attribute.Manufacturer `tag:"required"`
	Version      *attribute.Version
}

func (i Info) InstanceName() string {
	return "info"
}
