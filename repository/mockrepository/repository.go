package mockrepository

import (
	"context"

	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/repository"
)

type MockRepository struct {
	*MockActivityRepository
	*MockAnswerRepository
	*MockCampRepository
	*MockEventRepository
	*MockMessageRepository
	*MockOptionRepository
	*MockPaymentRepository
	*MockQuestionRepository
	*MockQuestionGroupRepository
	*MockRollCallRepository
	*MockRollCallReactionRepository
	*MockRoomRepository
	*MockRoomGroupRepository
	*MockUserRepository
}

func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	return &MockRepository{
		MockActivityRepository:         NewMockActivityRepository(ctrl),
		MockAnswerRepository:           NewMockAnswerRepository(ctrl),
		MockCampRepository:             NewMockCampRepository(ctrl),
		MockEventRepository:            NewMockEventRepository(ctrl),
		MockMessageRepository:          NewMockMessageRepository(ctrl),
		MockOptionRepository:           NewMockOptionRepository(ctrl),
		MockPaymentRepository:          NewMockPaymentRepository(ctrl),
		MockQuestionRepository:         NewMockQuestionRepository(ctrl),
		MockQuestionGroupRepository:    NewMockQuestionGroupRepository(ctrl),
		MockRollCallRepository:         NewMockRollCallRepository(ctrl),
		MockRollCallReactionRepository: NewMockRollCallReactionRepository(ctrl),
		MockRoomRepository:             NewMockRoomRepository(ctrl),
		MockRoomGroupRepository:        NewMockRoomGroupRepository(ctrl),
		MockUserRepository:             NewMockUserRepository(ctrl),
	}
}

func (m *MockRepository) Transaction(
	_ context.Context,
	fn func(tx repository.Repository) error,
) error {
	return fn(m)
}
