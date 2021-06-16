package utils

import (
	"strings"
	"unicode/utf8"

	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

func CheckDeviceName(name string) (err error) {

	inputName := "设备"
	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrapf(err, errors.NameNil, inputName)
		return
	}

	if utf8.RuneCountInString(name) > 20 {
		err = errors.Wrapf(err, errors.NameSizeLimit, inputName, 20)
		return
	}
	return
}
