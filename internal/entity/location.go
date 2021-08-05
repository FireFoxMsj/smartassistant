package entity

import (
	errors2 "errors"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"time"

	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// area_id与name复合唯一索引
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name" gorm:"uniqueIndex" `
	Sort int    `json:"sort" `

	CreatedAt time.Time `json:"created_at"`
}

func (d Location) TableName() string {
	return "locations"
}

func (d *Location) BeforeCreate(tx *gorm.DB) (err error) {
	// 房间名是否重复
	if LocationNameExist(d.Name) {
		err = errors.Wrap(err, status.LocationNameExist)
		return
	}

	// 添加房间, 新房间sort由后端生成
	var count int64
	if count, err = GetLocationCount(); err != nil {
		return
	} else {
		d.Sort = int(count) + 1
	}
	return
}

func CreateLocation(location *Location) error {
	return GetDB().Create(location).Error
}

func GetLocationByID(id int) (location Location, err error) {
	err = GetDB().First(&location, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.LocationNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func LocationNameExist(name string) bool {
	err := GetDB().First(&Location{}, "name = ?", name).Error
	return err == nil
}

func GetLocations() (locations []Location, err error) {

	err = GetDB().Order("sort asc").Find(&locations).Error
	return
}

func GetLocationCount() (count int64, err error) {
	err = GetDB().Model(&Location{}).Count(&count).Error
	return
}

func DelLocation(id int) (err error) {
	area := &Location{ID: id}
	err = GetDB().First(area).Delete(area).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.LocationNotExit)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func UpdateLocation(id int, updateLocation Location) (err error) {
	location := &Location{ID: id}
	err = GetDB().First(location).Updates(updateLocation).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New(status.LocationNotExit)
		} else {
			err = errors.New(errors.InternalServerErr)
		}
	}
	return
}

func EditLocationSort(id int, sort int) (err error) {
	err = GetDB().First(&Location{}, "id = ?", id).Update("sort", sort).Error
	return
}

func IsLocationExist(locationID int) bool {
	err := GetDB().First(&Location{}, "id = ?", locationID).Error
	return err == nil
}
