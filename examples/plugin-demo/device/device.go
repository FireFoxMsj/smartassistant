package device

import (
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

// LightBulbDevice 灯
type LightBulbDevice struct {
	LightBulb    instance.LightBulb
	Info0        instance.Info
	identity     string
	model        string
	manufacturer string
	ch           server.WatchChan
}

// SwitchDevice 三键开关
type SwitchDevice struct {
	Switch0      instance.Switch
	Switch1      instance.Switch
	Switch2      instance.Switch
	Info0        instance.Info
	identity     string
	model        string
	manufacturer string
	ch           server.WatchChan
}

func (sd *SwitchDevice) Identity() string {
	return sd.identity
}

func (sd *SwitchDevice) Info() server.DeviceInfo {
	return server.DeviceInfo{
		Identity:     sd.identity,
		Model:        sd.model,
		Manufacturer: sd.manufacturer,
	}
}

func (sd *SwitchDevice) update(instance string) attribute.UpdateFunc {
	return func(val interface{}) error {
		switch instance {
		case "switch0":
			sd.Switch0.Power.SetString(val.(string))
		case "switch1":
			sd.Switch1.Power.SetString(val.(string))
		case "switch2":
			sd.Switch2.Power.SetString(val.(string))
		}

		n := server.Notification{
			Identity:   sd.identity,
			InstanceID: 1,
			Attr:       "power",
			Val:        val,
		}
		select {
		case sd.ch <- n:
		default:
		}

		logger.Debug("notify:", n)

		return nil
	}
}

func (sd *SwitchDevice) Setup() error {
	sd.Info0.Identity.SetString(sd.identity)
	sd.Info0.Model.SetString(sd.model)
	sd.Info0.Manufacturer.SetString(sd.manufacturer)

	sd.Switch0.Power.SetString("off")
	sd.Switch1.Power.SetString("off")
	sd.Switch2.Power.SetString("off")

	//
	sd.Switch0.Power.SetUpdateFunc(sd.update("switch0"))
	sd.Switch0.Power.SetUpdateFunc(sd.update("switch1"))
	sd.Switch0.Power.SetUpdateFunc(sd.update("switch2"))

	sd.ch = make(server.WatchChan, 10)

	return nil
}

func (sd *SwitchDevice) Online() bool {
	return true
}

func (sd *SwitchDevice) Update() error {
	return nil
}

func (sd *SwitchDevice) Close() error {
	close(sd.ch)
	return nil
}

func (sd *SwitchDevice) GetChannel() server.WatchChan {
	return sd.ch
}

func NewSwitchDevice(identity string) *SwitchDevice {
	info := instance.Info{
		Name:         attribute.NewName(),
		Identity:     attribute.NewIdentity(),
		Model:        attribute.NewModel(),
		Manufacturer: attribute.NewManufacturer(),
		Version:      attribute.NewVersion(),
	}

	sd := SwitchDevice{
		Switch0:      instance.Switch{Power: attribute.NewPower()},
		Switch1:      instance.Switch{Power: attribute.NewPower()},
		Switch2:      instance.Switch{Power: attribute.NewPower()},
		Info0:        info,
		identity:     identity,
		model:        "M2",
		manufacturer: "zhiting",
		ch:           make(server.WatchChan, 5),
	}

	return &sd
}

func (d *LightBulbDevice) Online() bool {
	return true
}

func (d *LightBulbDevice) Identity() string {
	return d.identity
}

func (d *LightBulbDevice) update(attr string) attribute.UpdateFunc {
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

		logger.Debug("notify:", n)

		return nil
	}
}
func (d *LightBulbDevice) Setup() error {
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

func (d *LightBulbDevice) Update() error {
	// d.LightBulb.Power.SetString("on")
	// d.LightBulb.Brightness.SetInt(65)
	// d.LightBulb.ColorTemp.SetInt(5000)
	return nil
}

func (d *LightBulbDevice) Close() error {
	close(d.ch)
	return nil
}

func (d *LightBulbDevice) GetChannel() server.WatchChan {
	return d.ch
}

func (d LightBulbDevice) Info() server.DeviceInfo {
	return server.DeviceInfo{
		Identity:     d.identity,
		Model:        d.model,
		Manufacturer: d.manufacturer,
	}
}

func NewLightBulbDevice(identity string, model string) *LightBulbDevice {

	lightBulb := instance.LightBulb{
		Power:      attribute.NewPower(),
		ColorTemp:  instance.NewColorTemp(),
		Brightness: instance.NewBrightness(),
	}

	info := instance.Info{
		Name:         attribute.NewName(),
		Identity:     attribute.NewIdentity(),
		Model:        attribute.NewModel(),
		Manufacturer: attribute.NewManufacturer(),
		Version:      attribute.NewVersion(),
	}

	d := LightBulbDevice{
		LightBulb:    lightBulb,
		Info0:        info,
		identity:     identity,
		model:        model,
		manufacturer: "zhiting",
		ch:           make(chan server.Notification, 5),
	}
	return &d
}
