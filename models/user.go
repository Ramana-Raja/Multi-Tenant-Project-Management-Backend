package models

import (
	"gorm.io/gorm"
)

type User struct {
	Username    string      `json:"username" gorm:"unique"`
	Password    string      `json:"password"`
	Workspaces  []Workspace `gorm:"many2many:user_workspaces;"`
	Projects    []Project   `gorm:"many2many:project_members;"`
	LastLoginIP string
	gorm.Model
}
