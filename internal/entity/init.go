package entity

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/internal/config"

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
	sqldb, err := gorm.Open(sqlite.Open(config.GetConf().SmartAssistant.Db), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("数据库连接失败 %v", err.Error()))
	}
	// PRAGMA foreign_keys=ON 开启外键关联约束
	sqldb.Exec("PRAGMA foreign_keys=ON;")

	logMode := logger.Warn
	if config.GetConf().Debug {
		logMode = logger.Info
	}
	db = sqldb.Session(&gorm.Session{
		Logger: logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logMode,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})

	db.AutoMigrate(
		Device{}, Location{}, Area{}, Role{}, RolePermission{},
		User{}, UserRole{}, Scene{}, SceneCondition{},
		SceneTask{}, TaskLog{},
	)

}
