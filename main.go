package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/migration"
	gormRepository "github.com/traPtitech/rucQ/repository/gorm"
	"github.com/traPtitech/rucQ/router"
	"github.com/traPtitech/rucQ/service"
)

func main() {
	e := echo.New()

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${level}")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTokyo",
		user,
		password,
		host,
		port,
		database,
	)
	isDev := os.Getenv("RUCQ_ENV") == "development"
	gormLogLevel := logger.Silent

	if isDev {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(gormMysql.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(gormLogLevel),
		TranslateError: true,
	})

	if err != nil {
		e.Logger.Fatal(err)
	}

	if err := migration.Migrate(db); err != nil {
		e.Logger.Fatal(err)
	}

	if isDev {
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

	traqBaseURL := os.Getenv("TRAQ_API_BASE_URL")
	botAccessToken := os.Getenv("BOT_ACCESS_TOKEN")
	traqService := service.NewTraqService(traqBaseURL, botAccessToken)

	api.RegisterHandlers(e, router.NewServer(repo, traqService, isDev))
	e.Logger.Fatal(e.Start("0.0.0.0:8080"))

}
