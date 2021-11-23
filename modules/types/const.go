package types

const (
	SATokenKey      = "smart-assistant-token"
	SaModel         = "smart_assistant"
	ScopeTokenKey   = "scope-token"
	VerificationKey = "verification-code" // 临时密码

	RoleKey   = "role"
	OwnerRole = "owner"

	// 云端校验来自sa的请求时使用
	SAID  = "SA-ID"
	SAKey = "SA-Key"

	// 生成oauth token时使用
	AreaID  = "Area-ID"
	UserKey = "User-Key"

	DockerRegistry = "docker.yctc.tech"
)

const (
	CloudDisk     = "wangpan"
	CloudDiskAddr = "wangpan:8089"
)
