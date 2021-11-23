package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// WindowDoorSensor 门窗传感器
type WindowDoorSensor struct {
	WindowDoorClose *WindowDoorClose
	Battery *Battery
}

func (w WindowDoorSensor) InstanceName() string {
	return "window_door_sensor"
}

type WindowDoorClose struct {
	attribute.Int
}

func NewWindowDoorClose() *WindowDoorClose {
	return &WindowDoorClose{}
}