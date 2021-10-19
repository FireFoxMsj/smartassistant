package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRole(t *testing.T) {
	const testingRole = "testing_test_role"
	const anotherTestingRole = "another_testing_test_role"
	r, err := AddRole(testingRole)
	assert.Nil(t, err, "add role error: %v", err)
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, r.Name, testingRole)
	r, err = UpdateRole(r.ID, anotherTestingRole)
	assert.Nil(t, err, "update role error: %v", err)
	assert.Equal(t, r.Name, anotherTestingRole)
	err = DeleteRole(r.ID)
	assert.Nil(t, err, "delete role error: %v", err)
}
