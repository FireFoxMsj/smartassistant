package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与用户相关的响应状态码
const (
	AccountNotExistErr = iota + 5000
	AccountPassWordErr
	DelSelfErr
	UserNotExist
	AccountNameExist
	AccountNameInputNilErr
	AccountNameFormatErr
	NickNameInputNilErr
	NicknameLengthUpperLimit
	NicknameLengthLowerLimit
	PasswordInputNilErr
	PasswordFormatErr
	RequireLogin
	QRCodeInvalid
	QRCodeExpired
	QRCodeCreatorDeny
	GetQRCodeErr

	RoleNotExist
	RoleNameExist
	RoleNameInputNilErr
	RoleNameLengthLimit
	Deny
	NotAllowModifyRoleOfTheOwner
	OwnerForbidJoinAreaAgain

	GetUserTokenAuthDeny
	GetUserTokenDeny
)

func init() {
	errors.NewCode(AccountNotExistErr, "用户名不存在")
	errors.NewCode(AccountPassWordErr, "用户名或密码错误")
	errors.NewCode(DelSelfErr, "用户不能删除自己")
	errors.NewCode(UserNotExist, "用户不存在")
	errors.NewCode(AccountNameExist, "当前用户名已存在,请重新输入")
	errors.NewCode(AccountNameInputNilErr, "请输入用户名")
	errors.NewCode(AccountNameFormatErr, "用户名只能输入数字、字母、下划线，不能全部是数字")

	errors.NewCode(NickNameInputNilErr, "请输入昵称")
	errors.NewCode(NicknameLengthUpperLimit, "昵称长度不能大于20位")
	errors.NewCode(NicknameLengthLowerLimit, "昵称长度不能小于6为")
	errors.NewCode(PasswordInputNilErr, "请输入密码")
	errors.NewCode(PasswordFormatErr, "密码不能少于6位")
	errors.NewCode(RequireLogin, "用户未登录")
	errors.NewCode(QRCodeInvalid, "二维码无效")
	errors.NewCode(QRCodeExpired, "二维码已过期")
	errors.NewCode(QRCodeCreatorDeny, "二维码创建者无权限")
	errors.NewCode(GetQRCodeErr, "获取二维码失败")
	errors.NewCode(RoleNotExist, "该角色不存在")
	errors.NewCode(RoleNameExist, "该角色已存在,请重新输入")
	errors.NewCode(RoleNameInputNilErr, "请输入角色名称")
	errors.NewCode(RoleNameLengthLimit, "角色名称不能超过20位")
	errors.NewCode(Deny, "当前用户没有权限")
	errors.NewCode(NotAllowModifyRoleOfTheOwner, "不允许修改拥有者的角色")
	errors.NewCode(OwnerForbidJoinAreaAgain, "您是该家庭的拥有者，无需再次加入")

	errors.NewCode(GetUserTokenAuthDeny, "非法的认证token")
	errors.NewCode(GetUserTokenDeny, "不允许找回用户凭证")
}
