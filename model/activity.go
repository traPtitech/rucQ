package model

import "gorm.io/gorm"

type ActivityType string

const (
	ActivityTypeRoomCreated          ActivityType = "room_created"
	ActivityTypePaymentCreated       ActivityType = "payment_created"
	ActivityTypePaymentAmountChanged ActivityType = "payment_amount_changed"
	ActivityTypePaymentPaidChanged   ActivityType = "payment_paid_changed"
	ActivityTypeRollCallCreated      ActivityType = "roll_call_created"
	ActivityTypeQuestionCreated      ActivityType = "question_created"
)

type Activity struct {
	gorm.Model
	Type        ActivityType `gorm:"size:50;not null;index"`
	CampID      uint         `gorm:"not null"`
	UserID      *string      `gorm:"size:32"` // payment_* のみ使用
	User        *User        `gorm:"foreignKey:UserID;references:ID"`
	ReferenceID uint         `gorm:"not null"` // RoomID / PaymentID / RollCallID / QuestionGroupID
}
