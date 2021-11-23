package models

import (
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

func NewClientStore() *ClientStore {
	return &ClientStore{}
}

type ClientStore struct {
}

// GetByID 实现ClientStore接口的方法
func (c ClientStore) GetByID(ID string) (clientInfo oauth2.ClientInfo, err error) {
	info, err := entity.GetClientByClientID(ID)
	if err != nil {
		return
	}
	client := models.Client{
		ID:     info.ClientID,
		Secret: info.ClientSecret,
	}
	return &client, err
}
