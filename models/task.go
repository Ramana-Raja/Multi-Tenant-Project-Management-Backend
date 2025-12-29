package models

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	ProjectID   uint
	WorkspaceID string
	Title       string
	Status      string
	AssigneeID  *uint
}
