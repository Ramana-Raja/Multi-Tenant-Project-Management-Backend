package utils

import "multi-tenet/models"

func CreateAuditLog(userID uint, workspaceID uint, projectID *uint, taskID *uint, action string, details string) {
	log := models.AuditLog{
		UserID:      userID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		TaskID:      taskID,
		Action:      action,
		Details:     details,
	}

	models.DB.Create(&log)
}
