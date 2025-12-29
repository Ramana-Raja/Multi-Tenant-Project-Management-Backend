package controllers

import (
	"fmt"
	"multi-tenet/models"
	"multi-tenet/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Input_type struct {
	Name string `json:"name" binding:"required"`
}

func CreateWorkspace(c *gin.Context) {
	var input Input_type
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userIDVal, ok := c.Get("currentUser")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	userID := userIDVal.(uint)

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		workspace := models.Workspace{
			Name: input.Name,
		}

		if err := tx.Create(&workspace).Error; err != nil {
			return err
		}

		workspaceMember := models.WorkspaceMember{
			UserID:      userID,
			WorkspaceID: workspace.ID,

			Role:     "OWNER",
			JoinedAt: time.Now(),
		}

		if err := tx.Create(&workspaceMember).Error; err != nil {
			return err
		}

		utils.CreateAuditLog(userID, workspace.ID, nil, nil,
			"WORKSPACE_CREATED",
			fmt.Sprintf("Workspace '%s' created", workspace.Name),
		)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "workspace created"})
}

func Listworkspace(c *gin.Context) {
	userIDVal, ok := c.Get("currentUser")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	userID := userIDVal.(uint)

	var workspaces []models.Workspace

	err := models.DB.Joins("JOIN workspace_members wm ON wm.workspace_id = workspaces.id").Where("workspace.id = ?", userID).Find(workspaces).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workspaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
	})

}

func AddNewMember(c *gin.Context) {
	widParam := c.Param("wid")
	uidParam := c.Param("uid")

	wid64, err := strconv.ParseUint(widParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return
	}

	uid64, err := strconv.ParseUint(uidParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	workspaceID := uint(wid64)
	inviteUserID := uint(uid64)

	userIDVal, ok := c.Get("currentUser")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	var member models.WorkspaceMember
	if err := models.DB.
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).Error; err != nil {

		c.JSON(http.StatusForbidden, gin.H{"error": "you are not in this workspace"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, inviteUserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
		return
	}

	if member.Role != "OWNER" {
		c.JSON(http.StatusForbidden, gin.H{"error": "you dont have rights to add members"})
		return
	}

	var existing models.WorkspaceMember
	if err := models.DB.
		Where("workspace_id = ? AND user_id = ?", workspaceID, inviteUserID).
		First(&existing).Error; err == nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "user already in workspace"})
		return
	}
	newMember := models.WorkspaceMember{
		UserID:      inviteUserID,
		WorkspaceID: workspaceID,
		Role:        "MEMBER",
		JoinedAt:    time.Now(),
	}

	if err := models.DB.Create(&newMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
		return
	}
	utils.CreateAuditLog(userID, workspaceID, nil, nil,
		"MEMBER_ADDED",
		fmt.Sprintf("User %d added to workspace as %s", inviteUserID, "MEMBER"),
	)

	c.JSON(http.StatusOK, gin.H{
		"workspace_id": workspaceID,
		"member_id":    inviteUserID,
		"role":         newMember.Role,
	})
}

func UpdateRole(c *gin.Context) {
	widParam := c.Param("wid")
	uidParam := c.Param("uid")
	role := c.Param("role")

	wid64, err := strconv.ParseUint(widParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return
	}

	uid64, err := strconv.ParseUint(uidParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	workspaceID := uint(wid64)
	targetUserID := uint(uid64)

	userIDVal, ok := c.Get("currentUser")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uint)

	var member models.WorkspaceMember
	if err := models.DB.
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).Error; err != nil {

		c.JSON(http.StatusForbidden, gin.H{"error": "you are not in this workspace"})
		return
	}

	if member.Role != "OWNER" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only OWNER can update roles"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, targetUserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
		return
	}

	var targetMember models.WorkspaceMember
	if err := models.DB.
		Where("workspace_id = ? AND user_id = ?", workspaceID, targetUserID).
		First(&targetMember).Error; err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "user is not in this workspace"})
		return
	}

	if role != "ADMIN" && role != "MEMBER" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role type"})
		return
	}

	if err := models.DB.Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, targetUserID).
		Update("role", role).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	utils.CreateAuditLog(userID, workspaceID, nil, nil,
		"ROLE_UPDATED",
		fmt.Sprintf("User %d role changed to %s", targetUserID, role),
	)

	c.JSON(http.StatusOK, gin.H{
		"workspace_id": workspaceID,
		"user_id":      targetUserID,
		"role":         role,
	})
}
