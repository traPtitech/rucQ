package repository

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
}
