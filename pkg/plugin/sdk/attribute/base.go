package attribute

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type UpdateFunc func(val interface{}) error
type NotifyFunc func(val interface{}) error

type Setter interface {
	Set(val interface{}) error
}
type Notifier interface {
	Notify(val interface{}) error
	SetNotifyFunc(NotifyFunc)
}

type IntType interface {
	GetRange() (*int, *int)
	SetRange(int, int)
	GetInt() int
	SetInt(int)
}

type StringType interface {
	GetString() string
	SetString(string)
}

type BoolType interface {
	GetBool() bool
	SetBool(bool)
}

type Base struct {
	updateFn   UpdateFunc
	notifyFunc NotifyFunc
}

// SetUpdateFunc 设置属性更新函数
func (b *Base) SetUpdateFunc(fn UpdateFunc) {
	b.updateFn = fn
}

// Set 触发Base.updateFn，更新设备属性
func (b *Base) Set(val interface{}) error {
	if b.updateFn != nil {
		return b.updateFn(val)
	}
	logrus.Warn("update func not set")
	return nil
	return errors.New("update func not set")
}

// SetNotifyFunc  设置通知函数
func (b *Base) SetNotifyFunc(fn NotifyFunc) {
	b.notifyFunc = fn
}

// Notify 触发Base.notifyFn,通过channel通知SA
func (b *Base) Notify(val interface{}) error {
	if b.notifyFunc != nil {
		return b.notifyFunc(val)
	}
	logrus.Warn("notify func not set")
	return nil
	return errors.New("update func not set")
}

type Int struct {
	Base
	min, max *int
	v        int
}

func (i *Int) SetRange(min, max int) {
	if min > max {
		return
	}
	i.min = &min
	i.max = &max
}

func (i *Int) GetRange() (min, max *int) {
	return i.min, i.max
}

func (i *Int) SetInt(v int) {
	i.v = v
}

func (i *Int) GetInt() int {
	return i.v
}

type Bool struct {
	Base
	v bool
}

func (b *Bool) SetBool(v bool) {
	b.v = v
}

func (b *Bool) GetBool() bool {
	return b.v
}

type String struct {
	Base
	v           string
	validValues map[string]interface{}
}

type Enum struct {
	Base
	v     int
	enums map[int]struct{}
}

func (e *Enum) SetEnums(enums ...int) {
	if e.enums == nil {
		e.enums = make(map[int]struct{})
	}
	for i := range enums {
		e.enums[i] = struct{}{}
	}
}

func (e *Enum) GetEnum() int {
	return e.v
}

func (e *Enum) SetEnum(enum int) {
	e.v = enum
}

func (s *String) SetString(v string) {
	if len(s.validValues) != 0 {
		if _, ok := s.validValues[v]; !ok {
			logrus.Warning("invalid string value: ", v)
			// TODO return error
			return
		}
	}
	s.v = v
}
func (s String) GetString() string {
	return s.v
}
func StringWithValidValues(values ...string) String {
	s := String{}
	if len(s.validValues) == 0 {
		s.validValues = make(map[string]interface{})
	}
	for _, values := range values {
		s.validValues[values] = struct{}{}
	}
	return s
}
func TypeOf(iface interface{}) string {
	switch iface.(type) {
	case IntType:
		return "int"
	case BoolType:
		return "bool"
	case StringType:
		return "string"
	}
	return ""
}

func ValueOf(iface interface{}) interface{} {
	switch v := iface.(type) {
	case IntType:
		return v.GetInt()
	case BoolType:
		return v.GetBool()
	case StringType:
		return v.GetString()
	}
	return ""
}
