package orm

import (
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var once sync.Once

func GetDB() *gorm.DB {
	once.Do(func() {
		loadDB()
	})
	return db
}

func loadDB() {
	sqldb, err := gorm.Open(sqlite.Open("sadb.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("数据库连接失败 %v", err.Error()))
	}
	// PRAGMA foreign_keys=ON 开启外键关联约束
	sqldb.Exec("PRAGMA foreign_keys=ON;")

	db = sqldb.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Warn)})

	db.AutoMigrate(
		Device{}, Location{}, Area{}, Role{}, RolePermission{},
		User{}, UserRole{}, Scene{}, SceneCondition{}, ConditionItem{},
		SceneTask{}, SceneTaskDevice{}, TaskLog{},
	)

}
