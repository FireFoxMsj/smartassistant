package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateUser(t *testing.T) {
	ast := assert.New(t)

	var u = User{
		ID:          1,
		AccountName: "npc",
		Password:    "1234",
		Token:       "1234",
	}

	var errorTokenUser = User{
		ID:          2,
		AccountName: "npc",
		Password:    "1234",
		Token:       "1234",
	}

	err := CreateUser(&u)
	ast.NoError(err, "create user error: %v", err)

	err = CreateUser(&errorTokenUser)
	ast.Error(err, "create user error")
}

func TestGetUser(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999
	const existToken = "1234"
	const notExistToken = "9999"

	user, err := GetUserByID(existID)
	ast.NoError(err, "get user by id error: %v", err)
	ast.NotEmpty(user)

	user, err = GetUserByID(notExistID)
	ast.Error(err, "get user by id error")
	ast.Empty(user)

	user, err = GetUserByToken(existToken)
	ast.NoError(err, "get user by token error: %v", err)
	ast.NotEmpty(user)

	user, err = GetUserByToken(notExistToken)
	ast.Error(err, "get user by token error")
	ast.Empty(user)
}

func TestIsAccountNameExist(t *testing.T) {
	ast := assert.New(t)

	const existAccountName = "npc"
	const notExistAccountName = "9999"

	res := IsAccountNameExist(existAccountName)
	ast.True(res, "is account name exist error")

	res = IsAccountNameExist(notExistAccountName)
	ast.False(res, "is account name exist error")
}

func TestEditUser(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999
	const newAccountName = "ccc"

	var updateUser = User{
		ID:          1,
		AccountName: newAccountName,
		Password:    "1234",
		Token:       "1234",
	}

	err := EditUser(existID, updateUser)
	ast.NoError(err, "edit user error: %v", err)
	u, _ := GetUserByID(existID)
	ast.Equal(u.AccountName, newAccountName)

	err = EditUser(notExistID, updateUser)
	ast.Error(err, "edit user error")
}

func TestDelUser(t *testing.T) {
	ast := assert.New(t)

	const existID = 1
	const notExistID = 999

	err := DelUser(existID)
	ast.NoError(err, "delete user error: %v", err)
	u, _ := GetUserByID(existID)
	ast.Empty(u)

	err = DelUser(notExistID)
	ast.Error(err, "delete user error")
}
