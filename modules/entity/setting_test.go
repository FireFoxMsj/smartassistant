package entity

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type testStruct struct {
	A int
	B string
}

func TestAddSetting(t *testing.T) {

	var s string

	err := GetSetting("user_credential", &s)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	s = "aaaaaaaa"
	if err := UpdateSetting("user_credential", &s); err != nil {
		log.Fatalln(err)
	}
	var res string
	GetSetting("user_credential", &res)
	assert.Equal(t, s, res)

	s = "bbbbbbbb"
	if err := UpdateSetting("user_credential", &s); err != nil {
		log.Fatalln(err)
	}

	GetSetting("user_credential", &res)
	assert.Equal(t, s, res)

	var a = testStruct{123, "456"}
	if err := UpdateSetting("user_credential", &a); err != nil {
		log.Fatalln(err)
	}

	err = GetSetting("user_credential", &res)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)

	var b testStruct
	GetSetting("user_credential", &b)
	assert.Equal(t, a, b)

}
