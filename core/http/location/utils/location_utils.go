package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

func CheckLocationName(name string) (err error) {

	inputName := "房间"

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

func CheckLocationSort(sort int) (err error) {
	numReg := regexp.MustCompile(`^\d$`)
	if !numReg.MatchString(strconv.Itoa(sort)) {
		err = errors.Wrap(err, errors.BadRequest)
	}

	return
}
