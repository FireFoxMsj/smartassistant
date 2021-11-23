package entity

import (
	"gorm.io/datatypes"
	"time"
)

type DeviceStatusHistory struct {
	ID 			int				`json:"id"`
	DeviceID    int   	    	`json:"device_id" gorm:"uniqueIndex:area_id_mau_device"`
	Context  	datatypes.JSON  `json:"context"`
	UserID   	int				`json:"user_id"`
	AreaID      uint64  		`json:"area_id" gorm:"type:bigint;uniqueIndex:area_id_mau_device"`
	CreateAt    time.Time		`json:"create_at"`
}

func (d DeviceStatusHistory) TableName() string {
	return "device_status_history"
}

func CreateHistory(areaID uint64, deviceID, userID int, context []byte) (err error){
	history := &DeviceStatusHistory{
		DeviceID: deviceID,
		Context: context,
		UserID: userID,
		AreaID: areaID,
	}
	err = GetDB().Create(history).Error
	if err != nil {
		return
	}
	return
}

func GetDeviceHistoryByCondition(deviceID int, areaID uint64, ) (histories []DeviceStatusHistory, err error){
	err = GetDB().Find(&histories, "device_id=? and area_id=?", deviceID, areaID).Error
	return
}




