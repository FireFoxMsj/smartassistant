package entity

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	logger1 "github.com/zhiting-tech/smartassistant/pkg/logger"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var once sync.Once

var Tables []interface{} = []interface{}{
	Device{}, Location{}, Area{}, Role{}, RolePermission{},
	User{}, UserRole{}, Scene{}, SceneCondition{},
	SceneTask{}, TaskLog{}, GlobalSetting{}, PluginInfo{},
}

func GetDB() *gorm.DB {
	once.Do(func() {
		loadDB()
	})
	return db
}

func loadDB() {

	var dialect gorm.Dialector
	database := config.GetConf().SmartAssistant.Database
	driver := database.Driver
	switch driver {
	case "sqlite":
		dialect = sqlite.Open(filepath.Join(config.GetConf().SmartAssistant.DataPath(),
			"smartassistant", "sadb.db"))
	case "postgres", "postgresql":
		format := "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"
		dsn := fmt.Sprintf(format, database.Host, database.Port, database.Username,
			database.Password, database.Name, "disable")
		dialect = postgres.Open(dsn)
	default:
		panic(fmt.Errorf("invalid dialector %v", driver))
	}
	sqldb, err := gorm.Open(dialect, &gorm.Config{})
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

	if err = db.AutoMigrate(Tables...); err != nil {
		logger1.Panicf("migrate err:%s", err.Error())
	}
}

// FromArea 查找家庭对应的数据
func FromArea(areaID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("area_id=?", areaID)
	}
}

func GetDBWithAreaScope(areaID uint64) *gorm.DB {
	return GetDB().Scopes(FromArea(areaID))
}
