package mockrepository

import "go.uber.org/mock/gomock"

type MockRepository struct {
	*MockAnswerRepository
	*MockCampRepository
	*MockEventRepository
	*MockOptionRepository
	*MockPaymentRepository
	*MockQuestionRepository
	*MockQuestionGroupRepository
	*MockRoomRepository
	*MockRoomGroupRepository
	*MockUserRepository
}

func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	return &MockRepository{
		MockAnswerRepository:        NewMockAnswerRepository(ctrl),
		MockCampRepository:          NewMockCampRepository(ctrl),
		MockEventRepository:         NewMockEventRepository(ctrl),
		MockOptionRepository:        NewMockOptionRepository(ctrl),
		MockPaymentRepository:       NewMockPaymentRepository(ctrl),
		MockQuestionRepository:      NewMockQuestionRepository(ctrl),
		MockQuestionGroupRepository: NewMockQuestionGroupRepository(ctrl),
		MockRoomRepository:          NewMockRoomRepository(ctrl),
		MockRoomGroupRepository:     NewMockRoomGroupRepository(ctrl),
		MockUserRepository:          NewMockUserRepository(ctrl),
	}
}
