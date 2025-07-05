package gorm

import (
	"database/sql"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/traPtitech/rucQ/migration"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func setup(t *testing.T) *Repository {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "mariadb:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
			// MySQL側の設定でunexpected EOFを防ぐ
			"MYSQL_WAIT_TIMEOUT":                   "600",
			"MYSQL_INTERACTIVE_TIMEOUT":            "600",
			"MYSQL_NET_READ_TIMEOUT":               "30",
			"MYSQL_NET_WRITE_TIMEOUT":              "30",
			"MYSQL_MAX_CONNECTIONS":                "100",
			"MYSQL_MAX_CONNECT_ERRORS":             "10000",
			"MYSQL_CONNECT_TIMEOUT":                "10",
			"MYSQL_INNODB_BUFFER_POOL_SIZE":        "64M",
			"MYSQL_INNODB_LOG_FILE_SIZE":           "16M",
			"MYSQL_INNODB_FLUSH_LOG_AT_TRX_COMMIT": "2",
		},
		// より強固な健全性チェック設定
		WaitingFor: wait.ForAll(
			wait.ForLog("ready for connections").WithOccurrence(2), // 初期化と本起動の2回
			wait.ForListeningPort("3306/tcp"),
			wait.ForSQL("3306/tcp", "mysql", func(host string, port nat.Port) string {
				return "root:password@tcp(" + host + ":" + port.Port() + ")/database"
			}).WithQuery("SELECT 1").WithStartupTimeout(60*time.Second),
		).WithStartupTimeout(120 * time.Second),
	}
	container, err := testcontainers.GenericContainer(
		t.Context(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, container)
	})

	port, err := container.MappedPort(t.Context(), "3306")
	require.NoError(t, err)

	// データベース接続設定（unexpected EOF対策）
	loc, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	config := mysql.NewConfig()
	config.User = "root"
	config.Passwd = "password"
	config.Net = "tcp"
	config.Addr = "localhost:" + port.Port()
	config.DBName = "database"
	config.Collation = "utf8mb4_general_ci"
	config.ParseTime = true
	config.Loc = loc

	// 接続とタイムアウト設定を強化
	config.Timeout = 30 * time.Second      // 接続タイムアウト
	config.ReadTimeout = 30 * time.Second  // 読み取りタイムアウト
	config.WriteTimeout = 30 * time.Second // 書き込みタイムアウト

	// 接続の安定性向上のためのパラメータ
	config.Params = map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "True",
		"loc":       "Asia/Tokyo",
		// 接続エラーの処理を改善
		"interpolateParams": "true",
		"autocommit":        "true",
		// 接続の健全性チェック
		"checkConnLiveness": "true",
		"maxAllowedPacket":  "67108864", // 64MB
	}

	conn, err := sql.Open("mysql", config.FormatDSN())
	require.NoError(t, err)

	// 接続プールの設定を慎重に調整
	conn.SetMaxOpenConns(5)                   // 少し増やす
	conn.SetMaxIdleConns(2)                   // アイドル接続を維持
	conn.SetConnMaxLifetime(30 * time.Minute) // 接続の寿命を長く
	conn.SetConnMaxIdleTime(10 * time.Minute) // アイドル時間を設定

	t.Cleanup(func() {
		conn.Close()
	})

	db, err := gorm.Open(gormMysql.New(gormMysql.Config{
		Conn: conn,
	}), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = migration.Migrate(db)
	require.NoError(t, err)

	return NewGormRepository(db)
}

func mustCreateCamp(t *testing.T, r *Repository) model.Camp {
	t.Helper()

	dateStart := random.Time(t)
	dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
	camp := &model.Camp{
		DisplayID:          random.AlphaNumericString(t, 10),
		Name:               random.AlphaNumericString(t, 20),
		Guidebook:          random.AlphaNumericString(t, 100),
		IsDraft:            random.Bool(t),
		IsPaymentOpen:      random.Bool(t),
		IsRegistrationOpen: random.Bool(t),
		DateStart:          dateStart,
		DateEnd:            dateEnd,
	}
	err := r.CreateCamp(camp)

	require.NoError(t, err)

	// 時刻の精度などを揃えるため再取得する
	camp, err = r.GetCampByID(camp.ID)

	require.NoError(t, err)

	return *camp
}

func mustCreateEvent(t *testing.T, r *Repository, campID uint) model.Event {
	t.Helper()

	eventType := random.SelectFrom(
		t,
		model.EventTypeDuration,
		model.EventTypeMoment,
		model.EventTypeOfficial,
	)

	var event *model.Event

	switch eventType {
	case model.EventTypeDuration:
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		user := mustCreateUser(t, r)
		color := random.AlphaNumericString(t, 10)
		event = &model.Event{
			Type:         model.EventTypeDuration,
			Name:         random.AlphaNumericString(t, 20),
			Description:  random.AlphaNumericString(t, 100),
			Location:     random.AlphaNumericString(t, 50),
			TimeStart:    timeStart,
			TimeEnd:      &timeEnd,
			OrganizerID:  &user.ID,
			DisplayColor: &color,
			CampID:       campID,
		}

	case model.EventTypeMoment:
		event = &model.Event{
			Type:        model.EventTypeMoment,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   random.Time(t),
			CampID:      campID,
		}

	case model.EventTypeOfficial:
		timeStart := random.Time(t)
		timeEnd := timeStart.Add(time.Duration(random.PositiveInt(t)))
		event = &model.Event{
			Type:        model.EventTypeOfficial,
			Name:        random.AlphaNumericString(t, 20),
			Description: random.AlphaNumericString(t, 100),
			Location:    random.AlphaNumericString(t, 50),
			TimeStart:   timeStart,
			TimeEnd:     &timeEnd,
			CampID:      campID,
		}
	}

	err := r.CreateEvent(event)

	require.NoError(t, err)

	// 時刻の精度などを揃えるため再取得する
	event, err = r.GetEventByID(event.ID)

	require.NoError(t, err)

	return *event
}

func mustCreateQuestionGroup(t *testing.T, r *Repository, campID uint) model.QuestionGroup {
	t.Helper()

	questionGroup := &model.QuestionGroup{
		Name:        random.AlphaNumericString(t, 20),
		Description: random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
		Due:         random.Time(t),
		CampID:      campID,
	}

	err := r.CreateQuestionGroup(questionGroup)

	require.NoError(t, err)

	// 時刻の精度などを揃えるため再取得する
	questionGroup, err = r.GetQuestionGroup(questionGroup.ID)

	require.NoError(t, err)
	require.NotNil(t, questionGroup)

	return *questionGroup
}

func mustCreateQuestion(
	t *testing.T,
	r *Repository,
	questionGroupID uint,
	questionType model.QuestionType,
) model.Question {
	t.Helper()

	question := &model.Question{
		Type:            questionType,
		Title:           random.AlphaNumericString(t, 20),
		Description:     random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
		IsPublic:        random.Bool(t),
		IsOpen:          random.Bool(t),
		QuestionGroupID: questionGroupID,
	}

	switch questionType {
	case model.SingleChoiceQuestion, model.MultipleChoiceQuestion:
		question.Options = make([]model.Option, random.PositiveIntN(t, 10))

		for i := range question.Options {
			question.Options[i] = model.Option{
				Content: random.AlphaNumericString(t, 20),
			}
		}
	}

	err := r.CreateQuestion(question)

	require.NoError(t, err)

	// 時刻の精度などを揃えるため再取得する
	question, err = r.GetQuestionByID(question.ID)

	require.NoError(t, err)
	require.NotNil(t, question)

	return *question
}

func mustCreateUser(t *testing.T, r *Repository) model.User {
	t.Helper()

	userID := random.AlphaNumericString(t, 32)
	user, err := r.GetOrCreateUser(t.Context(), userID)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
	require.False(t, user.IsStaff)

	return *user
}
