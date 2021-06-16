package utils

import (
	"regexp"
)

// CheckIllegalRepeatDate 判断生效时间格式是否合法
func CheckIllegalRepeatDate(repeatDate string) bool {
	pattern := `^[1-7]{1,7}$`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(repeatDate) {
		return false
	}

	// "1122" 视为不合法字符串
	if !checkRepeatStr(repeatDate) {
		return false
	}
	return true
}

// checkRepeatStr 判断重复字符串
func checkRepeatStr(str string) bool {
	if str == "" {
		return false
	}

	var strMap = map[rune]bool{
		'1': false,
		'2': false,
		'3': false,
		'4': false,
		'5': false,
		'6': false,
		'7': false,
	}

	// 根据ASCII码判断字符是否重复
	for _, v := range str {
		if strMap[v] == true {
			return false
		}

		if val, ok := strMap[v]; ok && !val {
			strMap[v] = true
		}

	}
	return true
}
