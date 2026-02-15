package migration

import (
	"context"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

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
	CreatedAt time.Time
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

			// 1. Room → room_revealed
			//    RoomGroup経由でCampIDを取得
			var roomsWithCamp []v6RoomWithCampID
			if err := db.WithContext(ctx).
				Table("rooms").
				Select("rooms.id AS room_id, room_groups.camp_id AS camp_id, rooms.created_at").
				Joins("JOIN room_groups ON room_groups.id = rooms.room_group_id").
				Where("rooms.deleted_at IS NULL").
				Scan(&roomsWithCamp).Error; err != nil {
				return err
			}

			if len(roomsWithCamp) > 0 {
				activities := make([]v6Activity, 0, len(roomsWithCamp))
				for _, r := range roomsWithCamp {
					activities = append(activities, v6Activity{
						Model:       gorm.Model{CreatedAt: r.CreatedAt},
						Type:        "room_created",
						CampID:      r.CampID,
						ReferenceID: r.RoomID,
					})
				}

				if err := gorm.G[v6Activity](db).CreateInBatches(
					ctx,
					&activities,
					activityBatchSize,
				); err != nil {
					return err
				}
			}

			// 2. Payment → payment_created + payment_paid_changed
			if err := gorm.G[v6Payment](db).FindInBatches(ctx, activityBatchSize, func(data []v6Payment, _ int) error {
				if len(data) == 0 {
					return nil
				}

				activities := make([]v6Activity, 0, len(data)*2)
				for _, p := range data {
					userID := p.UserID
					activities = append(activities, v6Activity{
						Model:       gorm.Model{CreatedAt: p.CreatedAt},
						Type:        "payment_created",
						CampID:      p.CampID,
						UserID:      &userID,
						ReferenceID: p.ID,
					})
					if !p.UpdatedAt.Equal(p.CreatedAt) {
						activities = append(activities, v6Activity{
							Model:       gorm.Model{CreatedAt: p.UpdatedAt},
							Type:        "payment_paid_changed",
							CampID:      p.CampID,
							UserID:      &userID,
							ReferenceID: p.ID,
						})
					}
				}

				if len(activities) == 0 {
					return nil
				}

				return gorm.G[v6Activity](db).CreateInBatches(ctx, &activities, activityBatchSize)
			}); err != nil {
				return err
			}

			// 3. RollCall → roll_call_created
			if err := gorm.G[v6RollCall](db).FindInBatches(ctx, activityBatchSize, func(data []v6RollCall, _ int) error {
				if len(data) == 0 {
					return nil
				}

				activities := make([]v6Activity, 0, len(data))
				for _, rc := range data {
					activities = append(activities, v6Activity{
						Model:       gorm.Model{CreatedAt: rc.CreatedAt},
						Type:        "roll_call_created",
						CampID:      rc.CampID,
						ReferenceID: rc.ID,
					})
				}

				return gorm.G[v6Activity](db).CreateInBatches(ctx, &activities, activityBatchSize)
			}); err != nil {
				return err
			}

			// 4. QuestionGroup → question_created
			if err := gorm.G[v6QuestionGroup](db).FindInBatches(
				ctx,
				activityBatchSize,
				func(data []v6QuestionGroup, _ int) error {
					if len(data) == 0 {
						return nil
					}

					activities := make([]v6Activity, 0, len(data))
					for _, qg := range data {
						activities = append(activities, v6Activity{
							Model:       gorm.Model{CreatedAt: qg.CreatedAt},
							Type:        "question_created",
							CampID:      qg.CampID,
							ReferenceID: qg.ID,
						})
					}

					return gorm.G[v6Activity](db).CreateInBatches(ctx, &activities, activityBatchSize)
				},
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
