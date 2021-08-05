package device

import (
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

// Device fake device
type Device struct {
	LightBulb    instance.LightBulb
	Info0        instance.Info
	identity     string
	model        string
	manufacturer string
	ch           server.WatchChan
}

func (d *Device) Identity() string {
	return d.identity
}

func (d *Device) update(attr string) attribute.UpdateFunc {
	return func(val interface{}) error {
		switch attr {
		case "power":
			d.LightBulb.Power.SetString(val.(string))
		case "brightness":
			d.LightBulb.Brightness.SetInt(val.(int))
		case "color_temp":
			d.LightBulb.ColorTemp.SetInt(val.(int))
		}

		n := server.Notification{
			Identity:   d.identity,
			InstanceID: 1,
			Attr:       attr,
			Val:        val,
		}
		select {
		case d.ch <- n:
		default:
		}

		logrus.Debug("notify:", n)

		return nil
	}
}
func (d *Device) Setup() error {
	d.Info0.Identity.SetString(d.identity)
	d.Info0.Model.SetString(d.model)
	d.Info0.Manufacturer.SetString(d.manufacturer)

	d.LightBulb.Brightness.SetRange(1, 100)
	d.LightBulb.Power.SetString("off")
	switch d.model {
	case "lamp9":
		d.LightBulb.ColorTemp.SetRange(2700, 6500)
	case "ceiling17":
		d.LightBulb.ColorTemp.SetRange(3000, 5700)
	}

	// set up attribute updateFunc
	d.LightBulb.Power.SetUpdateFunc(d.update("power"))
	d.LightBulb.Brightness.SetUpdateFunc(d.update("brightness"))
	d.LightBulb.ColorTemp.SetUpdateFunc(d.update("color_temp"))

	d.ch = make(server.WatchChan, 10)
	return nil
}

func (d *Device) Update() error {
	// d.LightBulb.Power.SetString("on")
	// d.LightBulb.Brightness.SetInt(65)
	// d.LightBulb.ColorTemp.SetInt(5000)
	return nil
}

func (d *Device) Close() error {
	close(d.ch)
	return nil
}

func (d *Device) GetChannel() server.WatchChan {
	return d.ch
}

func (d Device) Info() server.DeviceInfo {
	return server.DeviceInfo{
		Identity:     d.identity,
		Model:        d.model,
		Manufacturer: d.manufacturer,
	}
}

func NewDemo(identity string) *Device {
	d := Device{
		LightBulb:    instance.NewLightBulb(),
		Info0:        instance.NewInfo(),
		identity:     identity,
		model:        "M1",
		manufacturer: "zhiting",
		ch:           make(chan server.Notification, 5),
	}
	return &d
}
