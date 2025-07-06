package main

import (
	"fmt"
	"os"
	"strings"

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

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	database := os.Getenv("DB_NAME")
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

	debug := os.Getenv("RUCQ_DEBUG") == "true"

	if debug {
		e.Use(middleware.CORS())
	} else {
		allowOrigins := strings.Split(os.Getenv("RUCQ_CORS_ALLOW_ORIGINS"), ",")

		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: allowOrigins,
		}))
	}

	// botがtraQからのイベントを受け取るエンドポイントを設定
	e.POST("/api/traq-events", router.TraqEventHandler)

	repo := gormRepository.NewGormRepository(db)

	api.RegisterHandlers(e, router.NewServer(repo, debug))
	e.Logger.Fatal(e.Start("0.0.0.0:8080"))

}
