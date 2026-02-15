//go:generate go tool mockgen -source=$GOFILE -destination=mockactivity/$GOFILE -package=mockactivity
package activity

import (
	"context"
	"time"

	"github.com/traPtitech/rucQ/model"
)

type ActivityService interface {
	// ユーザーに関連するアクティビティを取得し、付加情報を組み立てて返す
	GetActivities(ctx context.Context, campID uint, userID string) ([]ActivityResponse, error)

	// 各イベント発生時にルーターから呼ばれる
	RecordRoomCreated(ctx context.Context, room model.Room) error
	RecordRoomCreatedWithCampID(ctx context.Context, room model.Room, campID uint) error
	RecordPaymentCreated(ctx context.Context, payment model.Payment) error
	RecordPaymentAmountChanged(ctx context.Context, payment model.Payment) error
	RecordPaymentPaidChanged(ctx context.Context, payment model.Payment) error
	RecordRollCallCreated(ctx context.Context, rollCall model.RollCall) error
	RecordQuestionCreated(ctx context.Context, questionGroup model.QuestionGroup) error
}

// ActivityResponse はサービス層の応答型
type ActivityResponse struct {
	ID   uint
	Type model.ActivityType
	Time time.Time

	// type ごとの付加情報（該当するもののみ非nil）
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
