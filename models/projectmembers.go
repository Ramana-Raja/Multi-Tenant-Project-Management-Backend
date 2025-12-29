package models

import "time"

type ProjectMember struct {
	UserID    uint `gorm:"primaryKey"`
	ProjectID uint `gorm:"primaryKey"`

	User    User    `gorm:"foreignKey:UserID"`
	Project Project `gorm:"foreignKey:ProjectID"`

	Role     string
	JoinedAt time.Time
}
