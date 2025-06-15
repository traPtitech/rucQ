package model

import "gorm.io/gorm"

type RollCall struct {
	gorm.Model
	Name        string
	Description string
	Options     []string `gorm:"serializer:json"`
	Subjects    []User   `gorm:"many2many:roll_call_subjects;"`
}
