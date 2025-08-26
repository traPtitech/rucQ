package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v4Camp struct {
	gorm.Model
	RollCalls []v4RollCall `gorm:"foreignKey:CampID"`
}

func (v4Camp) TableName() string {
	return "camps"
}

type v4RollCall struct {
	gorm.Model
	CampID uint `gorm:"not null"`
}

func (v4RollCall) TableName() string {
	return "roll_calls"
}

func v4() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "4",
		Migrate: func(db *gorm.DB) error {
			if err := db.Migrator().AddColumn(&v4RollCall{}, "camp_id"); err != nil {
				return err
			}

			return db.Migrator().CreateConstraint(&v4Camp{}, "RollCalls")
		},
		Rollback: func(db *gorm.DB) error {
			if err := db.Migrator().DropConstraint(&v4Camp{}, "RollCalls"); err != nil {
				return err
			}

			return db.Migrator().DropColumn(&v4RollCall{}, "camp_id")
		},
	}
}
