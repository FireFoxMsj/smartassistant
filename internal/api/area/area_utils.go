package area

import (
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"strings"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func checkAreaName(name string) (err error) {

	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrap(err, status.AreaNameInputNilErr)
		return
	}

	if utf8.RuneCountInString(name) > 30 {
		err = errors.Wrap(err, status.AreaNameLengthLimit)
		return
	}
	return
}
