package config

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var dbFn string

// TestSetup 工具函数，用于初始化 config 测试环境
func TestSetup() {
	rand.Seed(time.Now().UnixNano())
	var re = regexp.MustCompile("/modules/.*")
	fn, _ := os.Getwd()
	fn, _ = filepath.Abs(fn)
	fn = re.ReplaceAllString(filepath.ToSlash(fn), "/app.test.yaml")
	fn = filepath.FromSlash(fn)
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		log.Println("请拷贝 app.yaml.example 到 app.test.yaml")
		log.Fatalf("test config file not exists: %v", fn)
	}
	InitConfig(fn)
	tmpDB := fmt.Sprintf("_%d.db", rand.Int())
	dbFn = strings.Replace(GetConf().SmartAssistant.Database.Name, ".db", tmpDB, 1)
	GetConf().SmartAssistant.Database.Name = dbFn
	log.Printf("using db file: %v", dbFn)
}

// TestTeardown 工具函数，测试完成后对 config 进行清理
func TestTeardown() {
	if _, err := os.Stat(dbFn); err == nil {
		log.Printf("cleaning up db file: %v", dbFn)
		_ = os.Remove(dbFn)
	}
}
