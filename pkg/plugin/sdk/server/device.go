package server

type DeviceInfo struct {
	Identity     string
	Model        string
	Manufacturer string
	AuthRequired bool
}

type Notification struct {
	Identity   string
	InstanceID int
	Attr       string
	Val        interface{}
}

type WatchChan chan Notification

type Device interface {
	Identity() string
	Info() DeviceInfo
	Setup() error
	Online() bool
	Update() error
	Close() error
	GetChannel() WatchChan
}

// AuthDevice 需要认证的设备
type AuthDevice interface {
	Device
	// IsAuth 返回设备是否成功认证/配对
	IsAuth() bool
	// Auth 认证/配对
	Auth(params map[string]string) error
	// RemoveAuthorization 取消认证/配对
	RemoveAuthorization(params map[string]string) error
}

type Notify struct {
	Identity   string    `json:"identity"`
	InstanceID int       `json:"instance_id"`
	Attribute  Attribute `json:"attribute"`
}
