package gorm

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/traPtitech/rucQ/migration"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/port"
	"github.com/traPtitech/rucQ/testutil/random"
)

// sanitizeStackIdentifier removes special characters from test names to create valid Docker Compose stack identifiers
func sanitizeStackIdentifier(testName string) string {
	// Replace all non-alphanumeric characters with hyphens (including underscores for Docker compatibility)
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	sanitized := re.ReplaceAllString(testName, "-")

	// Remove consecutive hyphens
	re = regexp.MustCompile(`-+`)
	sanitized = re.ReplaceAllString(sanitized, "-")

	// Trim leading/trailing hyphens and convert to lowercase
	sanitized = strings.Trim(sanitized, "-")
	sanitized = strings.ToLower(sanitized)

	// Ensure it starts with a letter, if not, prefix with "test"
	if len(sanitized) == 0 || !(sanitized[0] >= 'a' && sanitized[0] <= 'z') {
		sanitized = "test-" + sanitized
	}

	// Limit length to avoid overly long names
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

func setup(t *testing.T) *Repository {
	t.Helper()
	ctx := context.Background()

	// Generate random ports to avoid conflicts between parallel tests
	portNames := []string{"MARIADB_PORT", "RUCQ_PORT", "SWAGGER_PORT", "ADMINER_PORT", "TRAQ_CADDY_PORT", "TRAQ_SERVER_PORT"}
	randomPorts := port.MustGetFreePorts(len(portNames))
	portEnvMap := port.PortsToStringMap(portNames, randomPorts)

	// Create a compose stack using the root compose.yaml file with a unique stack identifier
	// This ensures each test gets its own set of containers with unique names/networks
	stackIdentifier := sanitizeStackIdentifier(fmt.Sprintf("test-%s-%d", t.Name(), rand.Int()))
	composeStack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../../compose.yaml"),
		compose.StackIdentifier(stackIdentifier),
	)
	require.NoError(t, err, "Failed to create compose stack")

	// Set random ports via environment variables
	composeWithEnv := composeStack.WithEnv(portEnvMap)

	t.Cleanup(func() {
		require.NoError(
			t,
			composeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal),
		)
	})

	// Configure wait strategy for mariadb service and start all services
	composeWithWait := composeWithEnv.WaitForService(
		"mariadb",
		wait.ForHealthCheck().WithStartupTimeout(60*time.Second),
	)
	err = composeWithWait.Up(ctx, compose.Wait(true))
	require.NoError(t, err, "Failed to start compose stack")

	// Stop unnecessary services to avoid port conflicts with other tests
	stopServices := []string{"rucq", "swagger", "adminer", "traq_caddy", "traq_server", "traq_ui"}
	for _, service := range stopServices {
		// Get service container and stop it (ignore errors if service doesn't exist or isn't running)
		if container, err := composeStack.ServiceContainer(ctx, service); err == nil {
			_ = container.Stop(ctx, nil)
		}
	}

	// Get MariaDB service container
	mariadbContainer, err := composeStack.ServiceContainer(ctx, "mariadb")
	require.NoError(t, err, "Failed to get MariaDB container")

	host, err := mariadbContainer.Host(ctx)
	require.NoError(t, err)
	port, err := mariadbContainer.MappedPort(ctx, "3306")
	require.NoError(t, err)

	loc, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	config := mysql.NewConfig()
	config.User = "root"
	config.Passwd = "password"
	config.Net = "tcp"
	config.Addr = fmt.Sprintf("%s:%s", host, port.Port())
	config.DBName = "rucq" // compose.yamlで定義したDB名'rucq'に合わせる
	config.Collation = "utf8mb4_general_ci"
	config.ParseTime = true
	config.Loc = loc

	db, err := gorm.Open(gormMysql.Open(config.FormatDSN()), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
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
