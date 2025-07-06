package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	gormMysql "gorm.io/driver/mysql"
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

	loc, err := time.LoadLocation("Asia/Tokyo")

	if err != nil {
		e.Logger.Fatal(err)
	}

	config := mysql.NewConfig()
	config.User = os.Getenv("DB_USER")
	config.Passwd = os.Getenv("DB_PASSWORD")
	config.Addr = fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	config.DBName = os.Getenv("DB_NAME")
	config.ParseTime = true
	config.Collation = "utf8mb4_general_ci"
	config.Loc = loc
	dsn := config.FormatDSN()
	db, err := gorm.Open(gormMysql.Open(dsn), &gorm.Config{
		TranslateError: true,
	})

	if err != nil {
		e.Logger.Fatal(err)
	}

	if err := migration.Migrate(db); err != nil {
		e.Logger.Fatal(err)
	}

	isDev := os.Getenv("RUCQ_ENV") == "development"

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

	api.RegisterHandlers(e, router.NewServer(repo, isDev))
	e.Logger.Fatal(e.Start("0.0.0.0:80"))

}
