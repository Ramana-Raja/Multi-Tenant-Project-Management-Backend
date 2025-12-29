package models

import "gorm.io/gorm"

type AuditLog struct {
	gorm.Model
	UserID      uint
	WorkspaceID uint
	ProjectID   *uint
	TaskID      *uint
	Action      string
	Details     string
}
