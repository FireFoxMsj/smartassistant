package errors

type Code struct {
	Status int    `json:"status"`
	Reason string `json:"reason"`
}

var codeMap = make(map[int]string)

func NewCode(status int, reason string) Code {
	if _, ok := codeMap[status]; ok {
		panic("status existed!")
	}
	codeMap[status] = reason
	return Code{status, reason}
}

var (
	OK                 = NewCode(0, "成功")
	InternalServerErr  = NewCode(1, "服务器异常")
	BadRequest         = NewCode(2, "错误请求")
	NotFound           = NewCode(3, "找不到资源")
	RoleNameExist      = NewCode(6, "该角色已存在，请重新输入")
	RoleNameNotExist   = NewCode(7, "该角色不存在")
	GetQRCodeErr       = NewCode(8, "获取二维码失败")
	DeviceExists       = NewCode(9, "设备已被添加")
	Deny               = NewCode(10, "没有权限")
	AccountNotExistErr = NewCode(11, "用户不存在")
	AccountPassWordErr = NewCode(12, "账户名或密码错误")
	RequireLogin       = NewCode(13, "用户未登录")
	QRCodeInvalid      = NewCode(14, "二维码无效")
	AreaNotExist       = NewCode(15, "家庭不存在")
	NotBoundDevice     = NewCode(16, "当前用户未绑定该设备")
	DataSyncFail       = NewCode(17, "数据同步失败,请重试")
	DelSelfErr         = NewCode(18, "用户不能删除自己")
	AlreadyBind        = NewCode(19, "SA已经被绑定")

	NameSizeLimit = NewCode(20, "%s名称长度不得超过%d")
	NameNil       = NewCode(21, "请输入%s名称")
	NameExist     = NewCode(22, "%s名已存在")

	LocationNotExit = NewCode(23, "房间不存在")
	AlreadyDataSync = NewCode(24, "数据已同步,禁止多次同步数据")
	UserNotExist    = NewCode(25, "用户不存在")
	CreatorQuitErr  = NewCode(26, "当前家庭创建者不允许退出家庭")
	RoleNotExist    = NewCode(27, "该角色不存在")

	InputSizeErr   = NewCode(28, "%s长度不能%s%d位")
	InputNilErr    = NewCode(29, "请输入%s")
	DeviceNotExist = NewCode(30, "设备不存在")

	AccountNameExist     = NewCode(31, "用户名已存在")
	AccountNameFormatErr = NewCode(32, "用户名只能输入数字、字母、下划线，不能全部是数字")
	PasswordFormatErr    = NewCode(33, "密码格式错误，仅支持输入数字、字母和符号， 且不少于6位")
	SceneNameExist       = NewCode(34, "与其他场景重名，请修改")

	ConditionMisMatchTypeAndConfigErr = NewCode(36, "场景触发条件类型与配置不一致")

	DeviceActionErr          = NewCode(39, "设备操作类型不存在")
	TaskTypeErr              = NewCode(40, "任务类型错误")
	DeviceOperationNotSetErr = NewCode(42, "设备操作未设置")

	QRCodeExpired            = NewCode(43, "二维码过期")
	QRCodeCreatorDeny        = NewCode(44, "二维码创建者无权限")
	SceneConditionNotExist   = NewCode(45, "场景触发条件不存在")
	SceneTaskNotExist        = NewCode(46, "场景执行任务不存在")
	SceneNotExist            = NewCode(47, "场景不存在")
	ParamIncorrectErr        = NewCode(48, "%s不正确")
	ConditionTimingCountErr  = NewCode(49, "定时触发条件只能添加一个")
	SceneCreateDeny          = NewCode(50, "您没有创建场景的权限")
	SceneDeleteDeny          = NewCode(51, "您没有删除场景的权限")
	DeviceOrSceneControlDeny = NewCode(52, "没有场景或设备的控制权限")
	SceneTypeForbidModify    = NewCode(53, "场景类型不允许修改")
	DeviceOffline            = NewCode(54, "设备断连")
)
