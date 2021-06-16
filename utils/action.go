package utils

const (
	ActionSwitch = "switch"
	ActionOnVal  = "on"  // 开
	ActionOffVal = "off" // 关

	ActionSetBright = "set_bright"

	ActionSetColorTemp = "set_color_temp"
)

var ActionMap = map[string]string{
	ActionSwitch:       ActionSwitch,
	ActionSetBright:    ActionSetBright,
	ActionSetColorTemp: ActionSetColorTemp,
}
