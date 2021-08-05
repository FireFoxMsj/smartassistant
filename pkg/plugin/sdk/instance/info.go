package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

type Info struct {
	Name         *attribute.Name         `tag:"identity:name;required"`
	Identity     *attribute.Identity     `tag:"identity:power;required"`
	Model        *attribute.Model        `tag:"name:model;required"`
	Manufacturer *attribute.Manufacturer `tag:"manufacturer:power;required"`
	Version      *attribute.Version      `tag:"name:version;required"`
	// FirmwareURL  *attribute.Version      `tag:"name:version;required"`
}

func (i Info) InstanceName() string {
	return "info"
}

func NewInfo() Info {
	return Info{
		Name:         &attribute.Name{},
		Identity:     &attribute.Identity{},
		Model:        &attribute.Model{},
		Manufacturer: &attribute.Manufacturer{},
		Version:      &attribute.Version{},
	}
}
