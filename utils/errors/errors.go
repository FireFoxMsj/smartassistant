package errors

import (
	"fmt"
	"github.com/pkg/errors"
)

type Error struct {
	Err  error
	Code Code
}

func (e Error) Error() string {
	return e.Code.Reason
}

// New
func New(code Code) error {
	return Newf(code)
}

// Newf
func Newf(code Code, args ...interface{}) error {

	if len(args) != 0 {
		code.Reason = fmt.Sprintf(code.Reason, args...)
	}
	return Error{
		Err:  errors.New(code.Reason),
		Code: code,
	}
}

// Wrap
func Wrap(err error, code Code) error {
	return Wrapf(err, code)
}

// WrapErrorf 支持格式化错误原因
func Wrapf(err error, code Code, args ...interface{}) error {
	switch v := err.(type) {
	case Error:
		err = v.Err
		code = v.Code
	default:
		err = errors.WithStack(err)
	}
	if len(args) != 0 {
		code.Reason = fmt.Sprintf(code.Reason, args...)
	}

	if err == nil { // 避免传nil error时日志没有任何调用栈
		err = errors.New(code.Reason)
	}

	return Error{
		Err:  err,
		Code: code,
	}
}

// Cause 获取原始错误
func Cause(err error) error {
	switch v := err.(type) {
	case Error:
		return errors.Cause(v.Err)
	default:
		return errors.Cause(err)
	}
}
