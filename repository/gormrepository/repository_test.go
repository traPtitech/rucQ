package gormrepository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/traPtitech/rucQ/migration"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

// GORMのログをテストケースごとに分けて出力するためのロガー
type testLogWriter struct {
	t *testing.T
}

// Printf implements the gorm.io/gorm/logger.Writer interface.
func (w *testLogWriter) Printf(format string, args ...any) {
	w.t.Logf(format, args...)
}

var containerAddr string

func TestMain(m *testing.M) {
	// クリーンアップはcomposeStack.Downで行われるためRyukはなくても問題ない
	err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	if err != nil {
		panic(fmt.Sprintf("failed to set environment variable: %v", err))
	}

	composeStack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../../compose.yaml"),
	)

	if err != nil {
		panic(fmt.Sprintf("failed to create compose stack: %v", err))
	}

	defer func() {
		err := composeStack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveImagesLocal,
			compose.RemoveVolumes(true),
		)

		if err != nil {
			panic(fmt.Sprintf("failed to stop compose stack: %v", err))
		}
	}()

	ctx := context.Background()

	if err := composeStack.
		WithEnv(map[string]string{
			"MARIADB_PORT": "0",
		}).
		WaitForService(
			"mariadb",
			wait.ForHealthCheck().WithStartupTimeout(60*time.Second),
		).
		Up(
			ctx,
			compose.Wait(true),
			compose.RunServices("mariadb"),
		); err != nil {
		panic(fmt.Sprintf("failed to start compose stack: %v", err))
	}

	mariadbContainer, err := composeStack.ServiceContainer(ctx, "mariadb")

	if err != nil {
		panic(fmt.Sprintf("failed to get MariaDB container: %v", err))
	}

	host, err := mariadbContainer.Host(ctx)

	if err != nil {
		panic(fmt.Sprintf("failed to get host: %v", err))
	}

	port, err := mariadbContainer.MappedPort(ctx, "3306")

	if err != nil {
		panic(fmt.Sprintf("failed to get mapped port: %v", err))
	}

	containerAddr = fmt.Sprintf("%s:%s", host, port.Port())

	m.Run()
}

func setup(t *testing.T) *Repository {
	t.Helper()

	loc, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	config := mysql.NewConfig()
	config.User = "root"
	config.Passwd = "password"
	config.Net = "tcp"
	config.Addr = containerAddr
	config.DBName = "rucq" // /dev/db/init.sqlで作成されるデータベース名
	config.Collation = "utf8mb4_general_ci"
	config.ParseTime = true
	config.Loc = loc
	// テストごとに異なるデータベース名を使用するため、1回接続してデータベースを作成する
	setupDB, err := sql.Open("mysql", config.FormatDSN())

	require.NoError(t, err)
	defer func() {
		err := setupDB.Close()

		require.NoError(t, err)
	}()

	// 一意な名前としてt.Name()を用いるが、長すぎる可能性があるため
	// ハッシュ化して適切な長さにする (SHA-256は16進数で64文字)
	h := sha256.New()

	h.Write([]byte(t.Name()))

	dbName := hex.EncodeToString(h.Sum(nil))
	_, err = setupDB.ExecContext(
		t.Context(),
		fmt.Sprintf("CREATE DATABASE `%s`", dbName),
	)

	require.NoError(t, err)

	config.DBName = dbName
	db, err := gorm.Open(gormMysql.Open(config.FormatDSN()), &gorm.Config{
		Logger: logger.New(
			&testLogWriter{t: t},
			logger.Config{Colorful: true, LogLevel: logger.Info},
		),
		TranslateError: true,
	})
	require.NoError(t, err)

	err = migration.Migrate(db)
	require.NoError(t, err)

	repo,err := NewGormRepository(db)
	
	if err != nil {
		log.Fatal(err)
	}

	return repo
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
	camp, err = r.GetCampByID(t.Context(), camp.ID)

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
	questionGroup, err = r.GetQuestionGroup(t.Context(), questionGroup.ID)

	require.NoError(t, err)
	require.NotNil(t, questionGroup)

	return *questionGroup
}

const maxOptions = 5

func mustCreateQuestion(
	t *testing.T,
	r *Repository,
	questionGroupID uint,
	questionType model.QuestionType,
	isPublic *bool,
) model.Question {
	t.Helper()

	publicValue := random.Bool(t)

	if isPublic != nil {
		publicValue = *isPublic
	}

	question := &model.Question{
		Type:            questionType,
		Title:           random.AlphaNumericString(t, 20),
		Description:     random.PtrOrNil(t, random.AlphaNumericString(t, 100)),
		IsPublic:        publicValue,
		IsOpen:          random.Bool(t),
		QuestionGroupID: questionGroupID,
	}

	switch questionType {
	case model.SingleChoiceQuestion, model.MultipleChoiceQuestion:
		// 2つ以上の選択肢を作成する
		question.Options = make([]model.Option, random.PositiveIntN(t, maxOptions)+1)

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

func mustCreatePayment(t *testing.T, r *Repository, userID string, campID uint) model.Payment {
	t.Helper()

	payment := model.Payment{
		Amount:     random.PositiveInt(t),
		AmountPaid: random.PositiveInt(t),
		UserID:     userID,
		CampID:     campID,
	}

	err := r.CreatePayment(t.Context(), &payment)

	require.NoError(t, err)
	require.NotZero(t, payment.ID)

	return payment
}

func mustCreateRoomGroup(t *testing.T, r *Repository, campID uint) *model.RoomGroup {
	t.Helper()

	roomGroup := &model.RoomGroup{
		Name:   random.AlphaNumericString(t, 20),
		CampID: campID,
	}

	err := r.CreateRoomGroup(t.Context(), roomGroup)

	require.NoError(t, err)
	require.NotNil(t, roomGroup)
	require.NotZero(t, roomGroup.ID)

	return roomGroup
}

func mustCreateRoom(
	t *testing.T,
	r *Repository,
	roomGroupID uint,
	members []model.User,
) *model.Room {
	t.Helper()

	room := &model.Room{
		Name:        random.AlphaNumericString(t, 20),
		RoomGroupID: roomGroupID,
		Members:     members,
	}

	err := r.CreateRoom(t.Context(), room)

	require.NoError(t, err)
	require.NotZero(t, room.ID)

	return room
}
