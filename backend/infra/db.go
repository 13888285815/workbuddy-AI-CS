package infra

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化数据库连接（从环境变量读取配置）
// 需要的环境变量：
// DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
func NewDB() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || name == "" {
		return nil, fmt.Errorf("database env not set: require DB_HOST, DB_PORT, DB_USER, DB_NAME")
	}

	// 尝试自动创建数据库（面向中小学生与零维护场景优化）
	dsnWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&charset=utf8mb4&loc=Local", user, password, host, port)
	dbRaw, err := gorm.Open(mysql.Open(dsnWithoutDB), &gorm.Config{})
	if err == nil {
		// 检查并创建数据库
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", name)
		_ = dbRaw.Exec(createDBSQL)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local", user, password, host, port, name)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
