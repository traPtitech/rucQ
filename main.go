package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/migration"
	"github.com/traPtitech/rucQ/repository/gormrepository"
	"github.com/traPtitech/rucQ/router"
	"github.com/traPtitech/rucQ/service/notification"
	"github.com/traPtitech/rucQ/service/scheduler"
	"github.com/traPtitech/rucQ/service/traq"
)

func main() {
	// TODO: graceful shutdownを実装する
	ctx := context.Background()
	e := echo.New()
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
		log.Fatal(err)
	}

	if err := migration.Migrate(db); err != nil {
		log.Fatal(err)
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError: true,
		LogError:    true,
		LogMethod:   true,
		LogStatus:   true,
		LogURI:      true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logLevel := slog.LevelInfo
			maxAttrs := 4
			attrs := make([]slog.Attr, 0, maxAttrs)
			attrs = append(
				attrs,
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
			)

			if v.Error != nil {
				logLevel = slog.LevelWarn

				if v.Status >= http.StatusInternalServerError {
					logLevel = slog.LevelError
				}

				var errorMessage any = v.Error.Error()
				var httpError *echo.HTTPError

				if errors.As(v.Error, &httpError) {
					errorMessage = httpError.Message

					if v.Status == http.StatusInternalServerError {
						errorMessage = httpError.Internal.Error()
					}
				}

				attrs = append(attrs, slog.Any("error", errorMessage))
			}

			slog.LogAttrs(
				c.Request().Context(),
				logLevel,
				"request",
				attrs...,
			)

			return nil
		},
	}))

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

	repo := gormrepository.NewGormRepository(db)

	traqBaseURL := cmp.Or(os.Getenv("TRAQ_API_BASE_URL"), "https://q.trap.jp/api/v3")
	botAccessToken := os.Getenv("TRAQ_BOT_ACCESS_TOKEN")
	traqService := traq.NewTraqService(traqBaseURL, botAccessToken)
	notificationService := notification.NewNotificationService(repo, traqService)
	schedulerService := scheduler.NewSchedulerService(repo, traqService)

	go schedulerService.Start(ctx)

	api.RegisterHandlers(e, router.NewServer(ctx, repo, notificationService, traqService, isDev))
	log.Fatal(e.Start("0.0.0.0:8080"))
}
