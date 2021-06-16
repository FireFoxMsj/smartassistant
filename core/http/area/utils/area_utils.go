package utils

import (
	"strings"
	"unicode/utf8"

	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

func CheckAreaName(name string) (err error) {

	inputName := "家庭"

	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrapf(err, errors.NameNil, inputName)
		return
	}

	if utf8.RuneCountInString(name) > 30 {
		err = errors.Wrapf(err, errors.NameSizeLimit, inputName, 30)
		return
	}
	return
}
