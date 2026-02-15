package migration

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// マイグレーション用の構造体定義

type v6Activity struct {
	gorm.Model
	Type        string `gorm:"size:50;not null;index"`
	CampID      uint   `gorm:"not null"`
	UserID      *string
	ReferenceID uint `gorm:"not null"`
}

func (v6Activity) TableName() string {
	return "activities"
}

// room_membersテーブルのJOIN用
type v6RoomWithCampID struct {
	RoomID    uint
	CampID    uint
	CreatedAt time.Time
}

type v6Payment struct {
	ID        uint `gorm:"primaryKey"`
	UserID    string
	CampID    uint
	UpdatedAt time.Time
}

func (v6Payment) TableName() string {
	return "payments"
}

type v6RollCall struct {
	ID        uint `gorm:"primaryKey"`
	CampID    uint
	CreatedAt time.Time
}

func (v6RollCall) TableName() string {
	return "roll_calls"
}

type v6QuestionGroup struct {
	ID        uint `gorm:"primaryKey"`
	CampID    uint
	CreatedAt time.Time
}

func (v6QuestionGroup) TableName() string {
	return "question_groups"
}

func v6() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "6",
		Migrate: func(db *gorm.DB) error {
			// activitiesテーブルを作成
			if err := db.AutoMigrate(&v6Activity{}); err != nil {
				return err
			}

			// 既存データからアクティビティを生成

			// 1. Room → room_revealed
			//    RoomGroup経由でCampIDを取得
			var roomsWithCamp []v6RoomWithCampID
			if err := db.
				Table("rooms").
				Select("rooms.id AS room_id, room_groups.camp_id AS camp_id, rooms.created_at").
				Joins("JOIN room_groups ON room_groups.id = rooms.room_group_id").
				Where("rooms.deleted_at IS NULL").
				Scan(&roomsWithCamp).Error; err != nil {
				return err
			}

			for _, r := range roomsWithCamp {
				activity := v6Activity{
					Model:       gorm.Model{CreatedAt: r.CreatedAt},
					Type:        "room_created",
					CampID:      r.CampID,
					ReferenceID: r.RoomID,
				}
				if err := db.Create(&activity).Error; err != nil {
					return err
				}
			}

			// 2. Payment → payment_amount_changed + payment_paid_changed
			var payments []v6Payment
			if err := db.Find(&payments).Error; err != nil {
				return err
			}

			for _, p := range payments {
				userID := p.UserID

				amountActivity := v6Activity{
					Model:       gorm.Model{CreatedAt: p.UpdatedAt},
					Type:        "payment_amount_changed",
					CampID:      p.CampID,
					UserID:      &userID,
					ReferenceID: p.ID,
				}
				if err := db.Create(&amountActivity).Error; err != nil {
					return err
				}

				paidActivity := v6Activity{
					Model:       gorm.Model{CreatedAt: p.UpdatedAt},
					Type:        "payment_paid_changed",
					CampID:      p.CampID,
					UserID:      &userID,
					ReferenceID: p.ID,
				}
				if err := db.Create(&paidActivity).Error; err != nil {
					return err
				}
			}

			// 3. RollCall → roll_call_created
			var rollCalls []v6RollCall
			if err := db.Find(&rollCalls).Error; err != nil {
				return err
			}

			for _, rc := range rollCalls {
				activity := v6Activity{
					Model:       gorm.Model{CreatedAt: rc.CreatedAt},
					Type:        "roll_call_created",
					CampID:      rc.CampID,
					ReferenceID: rc.ID,
				}
				if err := db.Create(&activity).Error; err != nil {
					return err
				}
			}

			// 4. QuestionGroup → question_created
			var questionGroups []v6QuestionGroup
			if err := db.Find(&questionGroups).Error; err != nil {
				return err
			}

			for _, qg := range questionGroups {
				activity := v6Activity{
					Model:       gorm.Model{CreatedAt: qg.CreatedAt},
					Type:        "question_created",
					CampID:      qg.CampID,
					ReferenceID: qg.ID,
				}
				if err := db.Create(&activity).Error; err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&v6Activity{})
		},
	}
}
