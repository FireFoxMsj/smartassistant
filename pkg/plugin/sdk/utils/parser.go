package utils

import (
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// Attribute represents a column of database
type Attribute struct {
	ID        int
	Model     interface{}
	TypeModel interface{}
	Name      string
	Type      string
	Tag       string
	Require   bool
}

func (a *Attribute) parseTag() {

	strs := strings.Split(a.Tag, ",")
	for _, v := range strs {
		infoStr := strings.Split(v, "=")
		if len(infoStr) != 0 {
			switch infoStr[0] {
			case "name":
			case "required":
				a.Require = true
			}
		}
	}
	return
}

type Instance struct {
	ID             int
	Model          interface{}
	Name           string
	Type           string
	Tag            string
	Attributes     []*Attribute
	AttributeNames []string
	attributeMap   map[string]*Attribute
}

// GetAttribute return instance by name
func (instance *Instance) GetAttribute(name string) *Attribute {
	return instance.attributeMap[name]
}

// Device represents a table of database
type Device struct {
	Model         interface{}
	Instances     []*Instance
	InstanceNames []string
	instanceMap   map[int]*Instance
}

// GetAttribute return Attribute by id
func (d *Device) GetAttribute(instanceID int, attr string) *Attribute {
	if instance, ok := d.instanceMap[instanceID]; ok {
		if attr, ok := instance.attributeMap[attr]; ok {
			return attr
		}
	}
	return nil
}

type DeviceName interface {
	DeviceName() string
}
type InstanceName interface {
	InstanceName() string
}

// Parse a struct to a Device instance
func Parse(dest interface{}) *Device {
	deviceType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	deviceValue := reflect.Indirect(reflect.ValueOf(dest))
	device := &Device{
		Model:       dest,
		instanceMap: make(map[int]*Instance),
	}
	instanceNum := 1
	for i := 0; i < deviceType.NumField(); i++ {
		p := deviceType.Field(i)
		v := reflect.Indirect(deviceValue.Field(i))
		if !p.Anonymous && ast.IsExported(p.Name) {
			if _, ok := v.Interface().(InstanceName); !ok {
				continue
			}
			instance := ParseInstance(v.Interface(), device)
			if v, ok := p.Tag.Lookup("tag"); ok {
				instance.Tag = v
			}
			instance.ID = instanceNum
			device.Instances = append(device.Instances, instance)
			device.InstanceNames = append(device.InstanceNames, p.Name)
			device.instanceMap[instance.ID] = instance
			instanceNum++
		}
	}
	return device
}

// ParseInstance a instance to a Instance instance
func ParseInstance(dest interface{}, device *Device) *Instance {
	instanceType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	instanceValue := reflect.Indirect(reflect.ValueOf(dest))
	var instanceName string
	t, ok := dest.(InstanceName)
	if !ok {
		instanceName = instanceType.Name()
	} else {
		instanceName = t.InstanceName()
	}
	instance := &Instance{
		Model:        dest,
		Name:         ToSnakeCase(instanceName),
		attributeMap: make(map[string]*Attribute),
		Type:         ToSnakeCase(instanceName),
	}

	attrNum := 1
	for i := 0; i < instanceType.NumField(); i++ {
		p := instanceType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			attr := Attribute{
				ID:    attrNum,
				Model: instanceValue.Field(i).Interface(),
				Name:  ToSnakeCase(p.Name),
				Type:  attribute.TypeOf(reflect.Indirect(reflect.New(p.Type)).Interface()), // FIXME
			}
			if v, ok := p.Tag.Lookup("tag"); ok {
				attr.Tag = v
			}
			attr.parseTag()
			// remove optional field
			if !attr.Require {
				if instanceValue.Field(i).IsZero() {
					continue
				}
				v := reflect.Indirect(instanceValue.Field(i))
				if vv, ok := v.Interface().(attribute.Base); ok {
					if !vv.Active {
						continue
					}
				}
			}
			instance.Attributes = append(instance.Attributes, &attr)
			instance.AttributeNames = append(instance.AttributeNames, p.Name)
			instance.attributeMap[ToSnakeCase(p.Name)] = &attr
			attrNum++
		}
	}
	return instance
}
