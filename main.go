package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/traP-jp/rucQ/backend/handler"
	"github.com/traP-jp/rucQ/backend/migration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	e := echo.New()

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${level}")
	}

	godotenv.Load(".env", "bot.env")

	user := os.Getenv("NS_MARIADB_USER")
	password := os.Getenv("NS_MARIADB_PASSWORD")
	host := os.Getenv("NS_MARIADB_HOSTNAME")
	database := os.Getenv("NS_MARIADB_DATABASE")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo", user, password, host, database)
	db, err := gorm.Open(mysql.Open(dsn))

	if err != nil {
		e.Logger.Fatal(err)
	}

	if err := migration.Migrate(db); err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173", // フロントエンド
			"http://localhost:8081", // Swagger UI
		},
	}))

	// botがtraQからのイベントを受け取るエンドポイントを設定
	e.POST("/api/traq-events", handler.TraqEventHandler)

	handler.RegisterHandlers(e, handler.NewServer(db))
	e.Logger.Fatal(e.Start(os.Getenv("RUCQ_BACKEND_ADDR")))

}
