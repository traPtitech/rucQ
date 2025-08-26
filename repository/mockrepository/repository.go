package mockrepository

import "go.uber.org/mock/gomock"

type MockRepository struct {
	*MockAnswerRepository
	*MockCampRepository
	*MockEventRepository
	*MockMessageRepository
	*MockOptionRepository
	*MockPaymentRepository
	*MockQuestionRepository
	*MockQuestionGroupRepository
	*MockRollCallRepository
	*MockRoomRepository
	*MockRoomGroupRepository
	*MockUserRepository
}

func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	return &MockRepository{
		MockAnswerRepository:        NewMockAnswerRepository(ctrl),
		MockCampRepository:          NewMockCampRepository(ctrl),
		MockEventRepository:         NewMockEventRepository(ctrl),
		MockMessageRepository:       NewMockMessageRepository(ctrl),
		MockOptionRepository:        NewMockOptionRepository(ctrl),
		MockPaymentRepository:       NewMockPaymentRepository(ctrl),
		MockQuestionRepository:      NewMockQuestionRepository(ctrl),
		MockQuestionGroupRepository: NewMockQuestionGroupRepository(ctrl),
		MockRollCallRepository:      NewMockRollCallRepository(ctrl),
		MockRoomRepository:          NewMockRoomRepository(ctrl),
		MockRoomGroupRepository:     NewMockRoomGroupRepository(ctrl),
		MockUserRepository:          NewMockUserRepository(ctrl),
	}
}
