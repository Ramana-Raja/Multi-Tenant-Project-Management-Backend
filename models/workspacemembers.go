package models

import "time"

type WorkspaceMember struct {
	UserID      uint `gorm:"primaryKey"`
	WorkspaceID uint `gorm:"primaryKey"`

	User      User
	Workspace Workspace

	Role     string
	JoinedAt time.Time
}
