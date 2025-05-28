package model

import "gorm.io/gorm"

type RollCall struct {
	gorm.Model
	Name        string
	Description string
	Options     []string `gorm:"serializer:json"`
	Subjects    []string `gorm:"serializer:json"`
}
