package orm

import (
	errors2 "errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

// Area 家庭/公司场景概念
type Area struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
}

func (d Area) TableName() string {
	return "areas"
}

func (d *Area) AfterDelete(tx *gorm.DB) (err error) {
	type TableInfo struct {
		TblName string `json:"tbl_name"`
	}
	var tis []TableInfo

	GetDB().Table("sqlite_master").Where("type = 'table'").Scan(&tis)

	for _, i := range tis {
		delSQL := fmt.Sprintf("DELETE FROM %s", i.TblName)
		if err = tx.Exec(delSQL).Error; err != nil {
			return
		}
	}
	return
}

func CreateArea(area *Area) error {
	return GetDB().Create(area).Error
}
func GetAreaByID(id int) (area Area, err error) {
	err = GetDB().First(&area, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.AreaNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetAreaCount() (count int64, err error) {
	err = GetDB().Model(Area{}).Count(&count).Error
	return
}

func GetAreas() (areas []Area, err error) {
	// 按照添加顺序获取(CreatedAt字段)
	err = GetDB().Order("created_at asc").Find(&areas).Error
	return

}

func DelAreaByID(id int) (err error) {
	s := Area{ID: id}
	err = GetDB().Delete(&s).Error
	return
}

// 修改Area名称后,同时需要修改location中旧名称
func EditAreaName(id int, name string) (err error) {
	err = GetDB().First(&Area{}, "id = ?", id).Update("name", name).Error
	return
}
