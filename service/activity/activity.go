//go:generate go tool mockgen -source=$GOFILE -destination=mockactivity/$GOFILE -package=mockactivity
package activity

import (
	"context"
	"time"

	"github.com/traPtitech/rucQ/model"
)

type ActivityService interface {
	GetActivities(ctx context.Context, campID uint, userID string) ([]ActivityResponse, error)
	RecordRoomCreated(ctx context.Context, room model.Room) error
	RecordPaymentCreated(ctx context.Context, payment model.Payment) error
	RecordPaymentAmountChanged(ctx context.Context, payment model.Payment) error
	RecordPaymentPaidChanged(ctx context.Context, payment model.Payment) error
	RecordRollCallCreated(ctx context.Context, rollCall model.RollCall) error
	RecordQuestionCreated(ctx context.Context, questionGroup model.QuestionGroup) error
}

type ActivityResponse struct {
	ID   uint
	Type model.ActivityType
	Time time.Time

	// typeごとの付加情報（該当するもののみ非nil）

	RoomCreated          *RoomCreatedDetail
	PaymentCreated       *PaymentCreatedDetail
	PaymentAmountChanged *PaymentChangedDetail
	PaymentPaidChanged   *PaymentChangedDetail
	RollCallCreated      *RollCallCreatedDetail
	QuestionCreated      *QuestionCreatedDetail
}

type RoomCreatedDetail struct{}

type PaymentCreatedDetail struct {
	Amount int
}

type PaymentChangedDetail struct {
	Amount int
}

type RollCallCreatedDetail struct {
	RollCallID uint
	Name       string
	IsSubject  bool
	Answered   bool
}

type QuestionCreatedDetail struct {
	QuestionGroupID uint
	Name            string
	Due             time.Time
	NeedsResponse   bool
}
