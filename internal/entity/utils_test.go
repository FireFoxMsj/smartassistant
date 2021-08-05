package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUtils(t *testing.T) {
	ast := assert.New(t)

	var utilsTest = []struct{
		RepeatDate string
		res bool
	}{
		{"1122", false},
		{"12", true},
		{"1233", false},
		{"", false},
		{"12345678", false},
	}

	for _, t := range utilsTest {
		res := CheckIllegalRepeatDate(t.RepeatDate)
		ast.Equal(t.res, res, "check repeatdate %v error", t.RepeatDate)
	}
}
