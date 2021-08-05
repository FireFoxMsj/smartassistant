package entity

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestCreateLocation(t *testing.T) {
	ast := assert.New(t)

	for i := 1; i < 11; i++ {
		location := Location{
			Name: "testLocation" + strconv.Itoa(i),
		}
		err := CreateLocation(&location)
		ast.NoError(err, "create location error: %v", err)
	}
	location := Location{
		Name: "testLocation1",
	}
	err := CreateLocation(&location)
	ast.Error(err, "create location error")
}

func TestLocationNameExist(t *testing.T) {
	ast := assert.New(t)

	const existName = "testLocation1"
	const notExistName = "666"

	tt := []struct{
		name string
		expectedRes bool
	}{
		{existName, true},
		{notExistName, false},
	}

	for _, t := range tt {
		ast.Equal(t.expectedRes, LocationNameExist(t.name))
	}
}

func TestGetLocationByID(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999

	location, err := GetLocationByID(existID)
	ast.NoError(err, "get location by id error: %v", err)
	ast.NotEmpty(location)

	location, err = GetLocationByID(notExistID)
	ast.Error(err, "get location by id error")
	ast.Empty(location)
}

func TestGetLocationCount(t *testing.T) {
	ast := assert.New(t)

	const correctCount int64= 10

	count, err := GetLocationCount()
	ast.NoError(err, "get location count error: %v", err)
	ast.Equal(correctCount, count)
}

func TestGetLocations(t *testing.T) {
	ast := assert.New(t)

	const correctCount= 10

	locations, err := GetLocations()
	ast.NoError(err, "get locations error: %v", err)
	ast.Equal(correctCount, len(locations))
}

func TestIsLocationExist(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999

	tt := []struct{
		id int
		expectedRes bool
	}{
		{existID, true},
		{notExistID, false},
	}

	for _, t := range tt {
		res := IsLocationExist(t.id)
		ast.Equal(t.expectedRes, res)
	}
}

func TestEditLocationSort(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999
	const newSort = 99

	err := EditLocationSort(existID, newSort)
	ast.NoError(err, "edit location sort error: %v", err)
	location, _ := GetLocationByID(existID)
	ast.Equal(newSort, location.Sort)

	err = EditLocationSort(notExistID, newSort)
	ast.Error(err, "edit location sort error")
}

func TestUpdateLocation(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999

	newLocation := Location{
		Name: "npcccc",
	}
	err := UpdateLocation(existID, newLocation)
	ast.NoError(err, "update location error: %v", err)

	err = UpdateLocation(notExistID, newLocation)
	ast.Error(err, "update location error")
}

func TestDelLocation(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999

	err := DelLocation(existID)
	ast.NoError(err, "delete location error: %v", err)

	err = DelLocation(notExistID)
	ast.Error(err, "delete location error")
}
