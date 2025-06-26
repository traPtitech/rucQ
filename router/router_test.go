package router

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	"github.com/traP-jp/rucQ/backend/api"
	"github.com/traP-jp/rucQ/backend/repository/mock"
)

type mockRepository struct {
	*mock.MockAnswerRepository
	*mock.MockCampRepository
	*mock.MockEventRepository
	*mock.MockOptionRepository
	*mock.MockQuestionRepository
	*mock.MockQuestionGroupRepository
	*mock.MockRoomRepository
	*mock.MockUserRepository
}

func newMockRepository(ctrl *gomock.Controller) *mockRepository {
	return &mockRepository{
		MockAnswerRepository:        mock.NewMockAnswerRepository(ctrl),
		MockCampRepository:          mock.NewMockCampRepository(ctrl),
		MockEventRepository:         mock.NewMockEventRepository(ctrl),
		MockOptionRepository:        mock.NewMockOptionRepository(ctrl),
		MockQuestionRepository:      mock.NewMockQuestionRepository(ctrl),
		MockQuestionGroupRepository: mock.NewMockQuestionGroupRepository(ctrl),
		MockRoomRepository:          mock.NewMockRoomRepository(ctrl),
		MockUserRepository:          mock.NewMockUserRepository(ctrl),
	}
}

type testHandler struct {
	repo   *mockRepository
	server *httptest.Server
}

func setup(t *testing.T) *testHandler {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := newMockRepository(ctrl)
	server := NewServer(repo, false)
	e := echo.New()

	api.RegisterHandlers(e, server)

	httptestServer := httptest.NewServer(e)

	t.Cleanup(func() {
		httptestServer.Close()
	})

	return &testHandler{
		repo:   repo,
		server: httptestServer,
	}
}
