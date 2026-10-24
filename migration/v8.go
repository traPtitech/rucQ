package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v8RollCallReaction struct {
	gorm.Model
	Content string
	UserID  string `gorm:"uniqueIndex:idx_user_id_roll_call_id"`

	RollCallID uint `gorm:"index;uniqueIndex:idx_user_id_roll_call_id"`
}

func (v8RollCallReaction) TableName() string {
	return "roll_call_reactions"
}

func v8() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "8",
		Migrate: func(db *gorm.DB) error {
			if err := db.Exec(`
				DELETE FROM roll_call_reactions
				WHERE deleted_at IS NOT NULL
			`).Error; err != nil {
				return err
			}

			if db.Migrator().HasIndex(&v8RollCallReaction{}, "idx_user_id_roll_call_id") {
				return nil
			}

			duplicateCleanerQuery := `
                DELETE t FROM roll_call_reactions t
                JOIN (
                    SELECT id,
                           ROW_NUMBER() OVER (
                               PARTITION BY roll_call_id, user_id
							   ORDER BY updated_at DESC, id DESC
                           ) as rn
                    FROM roll_call_reactions
                ) ranked ON t.id = ranked.id
                WHERE ranked.rn > 1;
            `

			if err := db.Exec(duplicateCleanerQuery).Error; err != nil {
				return err
			}

			// 重複が解消された後に、ユニークインデックスを追加する
			return db.Migrator().CreateIndex(&v8RollCallReaction{}, "idx_user_id_roll_call_id")
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropIndex(
				&v8RollCallReaction{},
				"idx_user_id_roll_call_id",
			)
		},
	}
}
