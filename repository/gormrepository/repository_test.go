package gormrepository

import (
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/migration"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

const initAndStartLogCount = 2

func setup(t *testing.T) *Repository {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "mariadb:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("ready for connections").WithOccurrence(initAndStartLogCount), // 初期化と本起動の2回
			wait.ForSQL("3306/tcp", "mysql", func(host string, port nat.Port) string {
				return fmt.Sprintf("root:password@tcp(%s:%s)/database", host, port.Port())
			}),
		),
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

	db, err := gorm.Open(gormMysql.Open(config.FormatDSN()), &gorm.Config{
		TranslateError: true,
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
		// 2つ以上の選択肢を作成する
		question.Options = make([]model.Option, random.PositiveIntN(t, 10)+1)

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
