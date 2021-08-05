package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	ast := assert.New(t)

	db := GetDB()
	ast.NotNil(db, "connect to database error")
}
