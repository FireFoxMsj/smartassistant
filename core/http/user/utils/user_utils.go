package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

var (
	NameSizeMin = 6
	NameSizeMax = 20
)

func CheckNickname(nickname string) (err error) {

	inputName := "昵称"

	if nickname == "" || strings.TrimSpace(nickname) == "" {
		err = errors.Wrapf(err, errors.InputNilErr, inputName)
		return
	}

	if utf8.RuneCountInString(nickname) > NameSizeMax {
		err = errors.Wrapf(err, errors.InputSizeErr, inputName, "大于", NameSizeMax)
		return
	}
	if utf8.RuneCountInString(nickname) < NameSizeMin {
		err = errors.Wrapf(err, errors.InputSizeErr, inputName, "少于", NameSizeMin)
		return
	}

	return
}

func CheckAccountName(accountName string) (err error) {
	inputName := "用户名"

	if accountName == "" || strings.TrimSpace(accountName) == "" {
		err = errors.Wrapf(err, errors.InputNilErr, inputName)
		return
	}

	if !CheckAccountNameFormat(accountName) {
		err = errors.New(errors.AccountNameFormatErr)
		return
	}

	if orm.IsAccountNameExist(accountName) {
		err = errors.New(errors.AccountNameExist)
		return
	}

	return
}

func CheckPassword(password string) (err error) {
	inputName := "密码"

	if password == "" || strings.TrimSpace(password) == "" {
		err = errors.Wrapf(err, errors.InputNilErr, inputName)
		return
	}

	if !CheckPasswordFormat(password) {
		err = errors.New(errors.PasswordFormatErr)
	}

	return
}

func CheckRoleID(roleID int) (err error) {
	if _, err = orm.GetRoleByID(roleID); err != nil {
		return
	}
	return
}

func CheckAccountNameFormat(accountName string) bool {

	namePattern := `^[\w]+$`
	allNumPattern := `^[\d]+$`
	nameReg := regexp.MustCompile(namePattern)
	allNumReg := regexp.MustCompile(allNumPattern)
	return nameReg.MatchString(accountName) && !allNumReg.MatchString(accountName)
}

func CheckPasswordFormat(password string) bool {
	pattern := `^[\x21-\x7e]{6,}$` // 字母和数字和符号的组合 大于六位
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(password)
}
