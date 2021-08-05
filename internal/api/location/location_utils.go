package location

import (
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func checkLocationName(name string) (err error) {

	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrap(err, status.LocationNameInputNilErr)
		return
	}

	if utf8.RuneCountInString(name) > 20 {
		err = errors.Wrap(err, status.LocationNameLengthLimit)
		return
	}

	return
}

func checkLocationSort(sort int) (err error) {
	numReg := regexp.MustCompile(`^\d$`)
	if !numReg.MatchString(strconv.Itoa(sort)) {
		err = errors.Wrap(err, errors.BadRequest)
	}

	return
}
