package router

import (
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/repository/mockrepository"
	"github.com/traPtitech/rucQ/service/mockservice"
)

type mockRepository struct {
	*mockrepository.MockAnswerRepository
	*mockrepository.MockCampRepository
	*mockrepository.MockEventRepository
	*mockrepository.MockOptionRepository
	*mockrepository.MockPaymentRepository
	*mockrepository.MockQuestionRepository
	*mockrepository.MockQuestionGroupRepository
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
		MockRoomRepository:          mockrepository.NewMockRoomRepository(ctrl),
		MockUserRepository:          mockrepository.NewMockUserRepository(ctrl),
	}
}

type testHandler struct {
	expect      *httpexpect.Expect
	repo        *mockRepository
	traqService *mockservice.MockTraqService
	server      *httptest.Server
}

func setup(t *testing.T) *testHandler {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := newMockRepository(ctrl)
	traqService := mockservice.NewMockTraqService(ctrl)
	server := NewServer(repo, traqService, false)
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
		expect:      expect,
		repo:        repo,
		traqService: traqService,
		server:      httptestServer,
	}
}
