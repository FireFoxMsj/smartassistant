package server

type DeviceInfo struct {
	Identity     string
	Model        string
	Manufacturer string
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
	Update() error
	Close() error
	GetChannel() WatchChan
}

type Notify struct {
	Identity   string    `json:"identity"`
	InstanceID int       `json:"instance_id"`
	Attribute  Attribute `json:"attribute"`
}
