package router

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/repository/mockrepository"
	"github.com/traPtitech/rucQ/service/notification/mocknotification"
	"github.com/traPtitech/rucQ/service/traq/mocktraq"
)

type testHandler struct {
	expect              *httpexpect.Expect
	repo                *mockrepository.MockRepository
	notificationService *mocknotification.MockNotificationService
	traqService         *mocktraq.MockTraqService
	// 基本的にはexpectを使うこと。
	// SSEなど、httpexpectでテストしづらいものをテストするときにのみ使用する
	e             *echo.Echo
	testServerURL string
}

func setup(t *testing.T) *testHandler {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := mockrepository.NewMockRepository(ctrl)
	traqService := mocktraq.NewMockTraqService(ctrl)
	notificationService := mocknotification.NewMockNotificationService(ctrl)
	server := NewServer(t.Context(), repo, notificationService, traqService, false)
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
		expect:              expect,
		repo:                repo,
		notificationService: notificationService,
		traqService:         traqService,
		e:                   e,
		testServerURL:       httptestServer.URL,
	}
}

func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(timeout):
		t.Error("timeout waiting for goroutine to finish")
	}
}
