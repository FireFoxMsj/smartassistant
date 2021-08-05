package rand

import (
	"math/rand"
	"time"
)

const (
	KindNum   = 1 << 0 // 纯数字
	KindLower = 1 << 1 // 小写字母
	KindUpper = 1 << 2 // 大写字母
	KindAll   = KindNum | KindLower | KindUpper
)

var randSource *rand.Rand

// 随机字符串
func StringK(size int, kind int) string {
	scope := make([][]int, 0, 3)
	if kind&KindNum != 0 {
		scope = append(scope, []int{10, 48})
	}
	if kind&KindLower != 0 {
		scope = append(scope, []int{26, 97})
	}
	if kind&KindUpper != 0 {
		scope = append(scope, []int{26, 65})
	}

	result := make([]byte, size)
	l := len(scope)
	for i := 0; i < size; i++ {
		index := randSource.Intn(l)
		s, base := scope[index][0], scope[index][1]
		result[i] = uint8(base + randSource.Intn(s))
	}
	return string(result)
}

func String(len int) string {
	return StringK(len, KindLower)
}

func init() {
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
}
