package migration

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/traP-jp/rucQ/backend/model"
	"gorm.io/gorm"
)

type v1OldBudget struct {
	gorm.Model
	UserID     uint
	CampID     uint
	Amount     *uint
	AmountPaid uint
}

func (v1OldBudget) TableName() string {
	return "budgets"
}

type v1NewBudget struct {
	gorm.Model
	UserTraqID string
	CampID     uint
	Amount     *uint
	AmountPaid uint
}

func (v1NewBudget) TableName() string {
	return "budgets"
}

type v1User struct {
	gorm.Model
	TraqID   string `gorm:"primaryKey"`
	IsStaff  bool   `gorm:"index"`
	Answers  []model.Answer
	TraqUuid string
}

func (v1User) TableName() string {
	return "users"
}

// 予算をGORMのIDではなくtraQ IDで管理するように変更
func v1() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1",
		Migrate: func(db *gorm.DB) error {
			if err := db.Migrator().AddColumn(&v1OldBudget{}, "user_traq_id"); err != nil {
				return fmt.Errorf("failed to add user_id column: %w", err)
			}

			var oldBudgets []v1OldBudget

			if err := db.Find(&oldBudgets).Error; err != nil {
				return fmt.Errorf("failed to get old budgets: %w", err)
			}

			for _, oldBudget := range oldBudgets {
				var user v1User

				if err := db.First(&user, oldBudget.UserID).Error; err != nil {
					return fmt.Errorf("failed to get user: %w", err)
				}

				if err := db.Model(&oldBudget).Update("user_traq_id", user.TraqID).Error; err != nil {
					return fmt.Errorf("failed to update user_traq_id: %w", err)
				}
			}

			if err := db.Migrator().DropColumn(&v1OldBudget{}, "user_id"); err != nil {
				return fmt.Errorf("failed to drop user_id column: %w", err)
			}

			return nil
		},
		Rollback: func(d *gorm.DB) error {
			if err := d.Migrator().AddColumn(&v1NewBudget{}, "user_id"); err != nil {
				return fmt.Errorf("failed to add user_id column: %w", err)
			}

			var newBudgets []v1NewBudget

			if err := d.Find(&newBudgets).Error; err != nil {
				return fmt.Errorf("failed to get old budgets: %w", err)
			}

			for _, oldBudget := range newBudgets {
				var user v1User

				if err := d.First(&user, oldBudget.UserTraqID).Error; err != nil {
					return fmt.Errorf("failed to get user: %w", err)
				}

				if err := d.Model(&oldBudget).Update("user_id", user.ID).Error; err != nil {
					return fmt.Errorf("failed to update user_id: %w", err)
				}
			}

			if err := d.Migrator().DropColumn(&v1OldBudget{}, "user_traq_id"); err != nil {
				return fmt.Errorf("failed to drop user_traq_id column: %w", err)
			}

			return nil
		},
	}
}
