package models

import "gorm.io/gorm"

type Workspace struct {
	gorm.Model
	Name    string
	Members []WorkspaceMember `gorm:"constraint:OnDelete:CASCADE;"`
}
