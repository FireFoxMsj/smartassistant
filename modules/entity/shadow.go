package entity

import (
	"errors"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type State struct {
	// Desired 期望值
	Desired map[int]map[string]interface{} `json:"desired"`
	// Reported 报告值
	Reported map[int]map[string]interface{} `json:"reported"`
}

type Metadata struct {
	Desired  map[int]map[string]AttrMetadata `json:"desired"`
	Reported map[int]map[string]AttrMetadata `json:"reported"`
}

type AttrMetadata struct {
	Timestamp int64 `json:"timestamp"`
}

// Shadow shadow of device
type Shadow struct {
	State     State    `json:"state"`
	Metadata  Metadata `json:"metadata"`
	Timestamp int      `json:"timestamp"`
	Version   int      `json:"version"`
}

func NewShadow() Shadow {
	return Shadow{
		State: State{
			Desired:  make(map[int]map[string]interface{}),
			Reported: make(map[int]map[string]interface{}),
		},
		Metadata: Metadata{
			Desired:  make(map[int]map[string]AttrMetadata),
			Reported: make(map[int]map[string]AttrMetadata),
		},
	}
}

func (s *Shadow) UpdateReported(instanceID int, attr server.Attribute) {

	if ins, ok := s.State.Reported[instanceID]; ok {
		ins[attr.Attribute] = attr.Val
	} else {
		s.State.Reported[instanceID] = map[string]interface{}{attr.Attribute: attr.Val}
	}
}

func (s Shadow) reportedAttr(instanceID int, attribute string) (val interface{}, err error) {
	if ins, ok := s.State.Reported[instanceID]; ok {

		if attr, ok := ins[attribute]; ok {
			return attr, nil
		}
	}
	err = errors.New("attr not found in shadow")
	return

}

func (s Shadow) Get(instanceID int, attribute string) (val interface{}, err error) {
	return s.reportedAttr(instanceID, attribute)
}
