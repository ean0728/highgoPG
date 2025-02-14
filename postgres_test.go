package highgoPG

import (
	"log"
	"testing"

	_ "github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDatabaseConnection(t *testing.T) {
	// dsn := "user=admin password=Hello@1234 dbname=tile_service host=192.168.0.227 port=7866 sslmode=disable"
	dsn := "postgresql://admin:Hello@1234@192.168.0.227:7866/tile_service?sslmode=disable"
	dialector := Open(dsn)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// 测试数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	log.Println("Database connection successful!")
} 