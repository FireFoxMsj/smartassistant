package types

const (
	EventSingleCast    = "single_cast"
	EventInstallPlugin = "plugin_installed"

	EventDeviceDiscovered = "device_discovered"
	EventStateChanged     = "state_changed"

	EventCallService = "call_service"
)

const (
	SATokenKey      = "smart-assistant-token"
	SaModel         = "smart_assistant"
	ScopeTokenKey   = "scope-token"
	VerificationKey = "verification-code"

	RoleKey   = "role"
	OwnerRole = "owner"

	// 云端校验来自sa的请求时使用
	SAID  = "SA-ID"
	SAKey = "SA-Key"
)

const (
	CloudDisk     = "wangpan"
	CloudDiskAddr = "wangpan:8089"
)
