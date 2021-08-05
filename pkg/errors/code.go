package errors

const (
	OK = iota
	InternalServerErr
	BadRequest
	NotFound
)

type Code struct {
	Status int    `json:"status"`
	Reason string `json:"reason"`
}

var codeMap = map[int]string{
	OK:                "成功",
	InternalServerErr: "服务器异常",
	BadRequest:        "错误请求",
	NotFound:          "找不到资源",
}

func NewCode(status int, reason string) {
	if _, ok := codeMap[status]; ok {
		panic("status existed!")
	}
	codeMap[status] = reason
}

func GetCodeReason(status int) string {
	return codeMap[status]
}

func GetCode(status int) Code {
	return Code{
		Status: status,
		Reason: GetCodeReason(status),
	}
}
