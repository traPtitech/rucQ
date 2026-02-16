package repository

import "context"

type Repository interface {
	ActivityRepository
	AnswerRepository
	CampRepository
	EventRepository
	MessageRepository
	OptionRepository
	PaymentRepository
	QuestionRepository
	QuestionGroupRepository
	RollCallRepository
	RollCallReactionRepository
	RoomGroupRepository
	RoomRepository
	UserRepository
	Transaction(ctx context.Context, fn func(tx Repository) error) error
}
