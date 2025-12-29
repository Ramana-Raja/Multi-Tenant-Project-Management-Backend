package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Name        string
	WorkspaceID uint

	Members   []User `gorm:"many2many:project_members;"`
	Tasks     []Task
	AuditLogs []AuditLog
}
