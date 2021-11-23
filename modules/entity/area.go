package entity

import (
	errors2 "errors"
	"gorm.io/gorm/schema"
	"reflect"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

const (
	AreaIDFieldName = "AreaID"
)

// Area 家庭
type Area struct {
	ID        uint64    `json:"id" gorm:"type:bigint"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	OwnerID   int       `json:"owner_id"`
	Deleted   gorm.DeletedAt
}

func (d Area) TableName() string {
	return "areas"
}

func (d *Area) AfterDelete(tx *gorm.DB) (err error) {
	areaTableName := d.TableName()
	// 遍历所有数据库表
	for _, table := range Tables {
		name, ok := table.(schema.Tabler)
		// 获取表名, 跳过自身
		if ok && areaTableName != name.TableName() {
			// 判断是否存在AreaID字段
			v := reflect.ValueOf(table)
			if v.FieldByName(AreaIDFieldName) == (reflect.Value{}) {
				continue
			}
			// 根据AreaID字段删除数据,判断是软删除还是硬删除
			if tx.Statement.Unscoped {
				err = tx.Model(table).Unscoped().Where("area_id = ?", d.ID).Delete(&table).Error
			} else {
				err = tx.Model(table).Where("area_id = ?", d.ID).Delete(&table).Error
			}

			if err != nil {
				return
			}
		}
	}
	return
}

func (d *Area) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = utils.SAAreaID()
	return nil
}

func CreateArea(name string) (area Area, err error) {
	if name != "" {
		area.Name = name
	}
	err = GetDB().Create(&area).Error
	if err != nil {
		return
	}
	return
}
func GetAreaByID(id uint64) (area Area, err error) {
	area, err = GetAreaResultById(id)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.AreaNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetAreaResultById(id uint64) (area Area, err error) {
	err = GetDB().First(&area, "id = ?", id).Error
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

func DelAreaByID(id uint64) (err error) {
	err = GetDB().Unscoped().Delete(&Area{ID: id}, id).Error
	return
}

// UpdateArea 修改Area名称后,同时需要修改location中旧名称
func UpdateArea(id uint64, name string) (err error) {
	err = GetDB().First(&Area{}, "id = ?", id).Update("name", name).Error
	return
}

func SetAreaOwnerID(id uint64, ownerID int, tx *gorm.DB) (err error) {
	err = tx.First(&Area{}, "id = ?", id).Update("owner_id", ownerID).Error
	return
}

// IsOwner 是否是area拥有者
func IsOwner(userID int) bool {
	var count int64
	GetDB().Model(&User{}).Where(User{ID: userID}).
		Joins("inner join areas on users.area_id=areas.id and areas.owner_id=users.id").
		Count(&count)
	return count > 0
}

// IsOwnerOfArea 是否是area拥有者
func IsOwnerOfArea(userID int, areaID uint64) bool {
	var count int64
	GetDB().Model(&User{}).Where(User{ID: userID}).
		Joins("inner join areas on users.area_id=areas.id and areas.owner_id=users.id").
		Where(Area{ID: areaID}).
		Count(&count)
	return count > 0
}

// GetAreaOwner 获取家庭的拥有者
func GetAreaOwner(areaID uint64) (user User, err error) {

	area, err := GetAreaByID(areaID)
	if err != nil {
		return
	}
	return GetUserByID(area.OwnerID)
}
