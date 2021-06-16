package components

import "gitlab.yctc.tech/root/smartassistent.git/core"

type Entity interface {
	Setup(m core.M) error
}

// ToggleEntity 开关定义
type ToggleEntity interface {
	Entity
	State(args core.M) error
	TurnOn(args core.M) error
	TurnOff(args core.M) error
	IsOn() error
}
