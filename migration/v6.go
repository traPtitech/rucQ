package migration

import (
	"context"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v6Activity struct {
	gorm.Model
	Type        string  `gorm:"size:50;not null;"`
	CampID      uint    `gorm:"not null"`
	Camp        *v6Camp `gorm:"foreignKey:CampID;references:ID;constraint:OnDelete:CASCADE"`
	UserID      *string `gorm:"size:32"`
	User        *v6User `gorm:"foreignKey:UserID;references:ID"`
	ReferenceID uint    `gorm:"not null"`
	Amount      *int
}

func (v6Activity) TableName() string {
	return "activities"
}

type v6User struct {
	ID string `gorm:"primaryKey;size:32"`
}

func (v6User) TableName() string {
	return "users"
}

type v6Camp struct {
	ID uint `gorm:"primaryKey"`
}

func (v6Camp) TableName() string {
	return "camps"
}

// room_membersテーブルのJOIN用
type v6RoomWithCampID struct {
	RoomID    uint
	CampID    uint
	CreatedAt time.Time
}

type v6Payment struct {
	gorm.Model
	Amount     int
	AmountPaid int
	UserID     string
	CampID     uint
}

func (v6Payment) TableName() string {
	return "payments"
}

type v6RollCall struct {
	gorm.Model
	CampID uint
}

func (v6RollCall) TableName() string {
	return "roll_calls"
}

type v6QuestionGroup struct {
	gorm.Model
	CampID uint
}

func (v6QuestionGroup) TableName() string {
	return "question_groups"
}

const activityBatchSize = 1000

func v6() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "6",
		Migrate: func(db *gorm.DB) error {
			ctx := context.Background()

			// activitiesテーブルを作成
			if err := db.Migrator().CreateTable(&v6Activity{}); err != nil {
				return err
			}

			// 既存データからアクティビティを生成

			// 1. Room → room_created
			//    RoomGroup経由でCampIDを取得
			var roomsWithCamp []v6RoomWithCampID
			if err := db.WithContext(ctx).
				Table("rooms").
				Select("rooms.id AS room_id, room_groups.camp_id AS camp_id, rooms.created_at").
				Joins("JOIN room_groups ON room_groups.id = rooms.room_group_id").
				Scan(&roomsWithCamp).Error; err != nil {
				return err
			}

			// 2. Payment → payment_created + payment_paid_changed
			payments, err := gorm.G[v6Payment](db).Find(ctx)
			if err != nil {
				return err
			}

			// 3. RollCall → roll_call_created
			rollCalls, err := gorm.G[v6RollCall](db).Find(ctx)
			if err != nil {
				return err
			}

			// 4. QuestionGroup → question_created
			questionGroups, err := gorm.G[v6QuestionGroup](db).Find(ctx)
			if err != nil {
				return err
			}

			updatedPaymentsCount := 0
			for _, p := range payments {
				if !p.UpdatedAt.Equal(p.CreatedAt) {
					updatedPaymentsCount++
				}
			}

			activitiesCapacity := len(
				roomsWithCamp,
			) + len(
				rollCalls,
			) + len(
				questionGroups,
			) + len(
				payments,
			) + updatedPaymentsCount
			activities := make([]v6Activity, 0, activitiesCapacity)

			for _, r := range roomsWithCamp {
				activities = append(activities, v6Activity{
					Model:       gorm.Model{CreatedAt: r.CreatedAt},
					Type:        "room_created",
					CampID:      r.CampID,
					ReferenceID: r.RoomID,
				})
			}

			for _, p := range payments {
				userID := p.UserID
				amount := p.Amount
				amountPaid := p.AmountPaid
				activities = append(activities, v6Activity{
					Model:       gorm.Model{CreatedAt: p.CreatedAt},
					Type:        "payment_created",
					CampID:      p.CampID,
					UserID:      &userID,
					ReferenceID: p.ID,
					Amount:      &amount,
				})
				if !p.UpdatedAt.Equal(p.CreatedAt) {
					// 何が更新されたか正確には分からないが、大抵の場合支払済み金額が
					// 更新されるのでpayment_paid_changedとする
					activities = append(activities, v6Activity{
						Model:       gorm.Model{CreatedAt: p.UpdatedAt},
						Type:        "payment_paid_changed",
						CampID:      p.CampID,
						UserID:      &userID,
						ReferenceID: p.ID,
						Amount:      &amountPaid,
					})
				}
			}

			for _, rc := range rollCalls {
				activities = append(activities, v6Activity{
					Model:       gorm.Model{CreatedAt: rc.CreatedAt},
					Type:        "roll_call_created",
					CampID:      rc.CampID,
					ReferenceID: rc.ID,
				})
			}

			for _, qg := range questionGroups {
				activities = append(activities, v6Activity{
					Model:       gorm.Model{CreatedAt: qg.CreatedAt},
					Type:        "question_created",
					CampID:      qg.CampID,
					ReferenceID: qg.ID,
				})
			}

			if len(activities) == 0 {
				return nil
			}

			if err := gorm.G[v6Activity](db).CreateInBatches(
				ctx,
				&activities,
				activityBatchSize,
			); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&v6Activity{})
		},
	}
}
