package router

import (
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/repository/mockrepository"
)

type mockRepository struct {
	*mockrepository.MockAnswerRepository
	*mockrepository.MockCampRepository
	*mockrepository.MockEventRepository
	*mockrepository.MockOptionRepository
	*mockrepository.MockPaymentRepository
	*mockrepository.MockQuestionRepository
	*mockrepository.MockQuestionGroupRepository
	*mockrepository.MockRoomGroupRepository
	*mockrepository.MockRoomRepository
	*mockrepository.MockUserRepository
}

func newMockRepository(ctrl *gomock.Controller) *mockRepository {
	return &mockRepository{
		MockAnswerRepository:        mockrepository.NewMockAnswerRepository(ctrl),
		MockCampRepository:          mockrepository.NewMockCampRepository(ctrl),
		MockEventRepository:         mockrepository.NewMockEventRepository(ctrl),
		MockOptionRepository:        mockrepository.NewMockOptionRepository(ctrl),
		MockPaymentRepository:       mockrepository.NewMockPaymentRepository(ctrl),
		MockQuestionRepository:      mockrepository.NewMockQuestionRepository(ctrl),
		MockQuestionGroupRepository: mockrepository.NewMockQuestionGroupRepository(ctrl),
		MockRoomGroupRepository:     mockrepository.NewMockRoomGroupRepository(ctrl),
		MockRoomRepository:          mockrepository.NewMockRoomRepository(ctrl),
		MockUserRepository:          mockrepository.NewMockUserRepository(ctrl),
	}
}

type testHandler struct {
	expect *httpexpect.Expect
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

	expect := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  httptestServer.URL,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	return &testHandler{
		expect: expect,
		repo:   repo,
		server: httptestServer,
	}
}
