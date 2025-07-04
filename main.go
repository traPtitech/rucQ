package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/migration"
	gormRepository "github.com/traPtitech/rucQ/repository/gorm"
	"github.com/traPtitech/rucQ/router"
)

func main() {
	e := echo.New()

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${level}")
	}

	//nolint:errcheck // 開発環境でしか使用しないため、エラーは無視
	godotenv.Load(".env", "bot.env")

	user := os.Getenv("NS_MARIADB_USER")
	password := os.Getenv("NS_MARIADB_PASSWORD")
	host := os.Getenv("NS_MARIADB_HOSTNAME")
	database := os.Getenv("NS_MARIADB_DATABASE")
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo",
		user,
		password,
		host,
		database,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})

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
	e.POST("/api/traq-events", router.TraqEventHandler)

	debug := os.Getenv("RUCQ_DEBUG") == "true"
	repo := gormRepository.NewGormRepository(db)

	api.RegisterHandlers(e, router.NewServer(repo, debug))
	e.Logger.Fatal(e.Start(os.Getenv("RUCQ_BACKEND_ADDR")))

}
