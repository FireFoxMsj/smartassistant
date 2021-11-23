package instance

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"

// SecuritySystem 安全系统
type SecuritySystem struct {
	TargetState *TargetState
	CurrentState *CurrentState
}

func (w SecuritySystem) InstanceName() string {
	return "security_system"
}

// TargetState 目标状态
type TargetState struct {
	attribute.Int
}

func NewTargetState() *TargetState {
	return &TargetState{}
}

// CurrentState 当前状态
type CurrentState struct {
	attribute.Int
}

func NewCurrentState() *CurrentState {
	return &CurrentState{}
}
