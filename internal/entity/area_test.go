package entity

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateArea(t *testing.T) {
	ast := assert.New(t)

	for i := 1; i < 11; i++ {
		area := Area{
			Name: "testArea" + strconv.Itoa(i),
		}
		err := CreateArea(&area)
		ast.NoError(err, "create area error: %v", err)
	}

	area := Area{
		Name: "testArea1",
	}
	err := CreateArea(&area)
	ast.Error(err, "create area error")
}

func TestGetAreaByID(t *testing.T) {
	ast := assert.New(t)

	for i := 1; i < 11; i++ {
		area, err := GetAreaByID(i)
		ast.NoError(err, "get area error: %v", err)
		ast.NotEmpty(area)
	}
}

func TestGetAreaCount(t *testing.T) {
	ast := assert.New(t)

	const correctCount int64 = 10
	count, err := GetAreaCount()
	ast.NoError(err, "get area count error: %v", err)
	ast.Equal(correctCount, count)
}

func TestGetAreas(t *testing.T) {
	ast := assert.New(t)

	const correctCount = 10

	areas, err := GetAreas()
	ast.NoError(err, "get areas error: %v", err)
	ast.Equal(len(areas), correctCount)
}

func TestUpdateArea(t *testing.T) {
	ast := assert.New(t)

	const newName = "new"
	const existID = 1
	const notExistID = 999

	err := UpdateArea(existID, newName)
	ast.NoError(err, "update area error: %v", err)

	err = UpdateArea(notExistID, newName)
	ast.Error(err, "update area error")
}

func TestDelAreaByID(t *testing.T) {
	ast := assert.New(t)

	const notExistID = 999

	for i := 1; i < 11; i++ {
		err := DelAreaByID(i)
		ast.NoError(err, "delete area error: %v", err)
	}

	err := DelAreaByID(notExistID)
	ast.NoError(err, "delete area error")
}
