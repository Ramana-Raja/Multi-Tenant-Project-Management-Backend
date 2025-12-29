package controllers

import (
	"fmt"
	"multi-tenet/models"
	"multi-tenet/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func requireWorkspaceMember(workspaceID uint, userID uint, roles ...string) (bool, string) {
	var member models.WorkspaceMember
	if err := models.DB.
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).Error; err != nil {
		return false, ""
	}

	if len(roles) == 0 {
		return true, member.Role
	}

	for _, r := range roles {
		if member.Role == r {
			return true, member.Role
		}
	}

	return false, member.Role
}

func CreateProject(c *gin.Context) {
	wid64, _ := strconv.ParseUint(c.Param("wid"), 10, 64)
	workspaceID := uint(wid64)

	userID := c.MustGet("currentUser").(uint)

	ok, _ := requireWorkspaceMember(workspaceID, userID, "OWNER")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := models.Project{
		Name:        input.Name,
		WorkspaceID: workspaceID,
	}

	if err := models.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	utils.CreateAuditLog(userID, workspaceID, &project.ID, nil,
		"PROJECT_CREATED",
		fmt.Sprintf("Project '%s' created", project.Name),
	)

	c.JSON(http.StatusCreated, gin.H{"project": project})
}

func GetProjects(c *gin.Context) {
	wid64, _ := strconv.ParseUint(c.Param("wid"), 10, 64)
	workspaceID := uint(wid64)

	userID := c.MustGet("currentUser").(uint)

	ok, _ := requireWorkspaceMember(workspaceID, userID)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}

	var projects []models.Project
	if err := models.DB.Where("workspace_id = ?", workspaceID).Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func CreateTask(c *gin.Context) {
	wid64, _ := strconv.ParseUint(c.Param("wid"), 10, 64)
	pid64, _ := strconv.ParseUint(c.Param("pid"), 10, 64)

	workspaceID := uint(wid64)
	projectID := uint(pid64)
	userID := c.MustGet("currentUser").(uint)

	ok, _ := requireWorkspaceMember(workspaceID, userID, "OWNER")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}

	var input struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := models.Task{
		Title:     input.Title,
		Status:    "TODO",
		ProjectID: projectID,
	}

	if err := models.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	utils.CreateAuditLog(userID, workspaceID, &projectID, &task.ID,
		"TASK_CREATED",
		fmt.Sprintf("Task '%s' created", task.Title),
	)

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

func GetTasks(c *gin.Context) {
	wid64, _ := strconv.ParseUint(c.Param("wid"), 10, 64)
	pid64, _ := strconv.ParseUint(c.Param("pid"), 10, 64)

	workspaceID := uint(wid64)
	projectID := uint(pid64)
	userID := c.MustGet("currentUser").(uint)

	ok, _ := requireWorkspaceMember(workspaceID, userID)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}

	var tasks []models.Task
	if err := models.DB.Where("project_id = ?", projectID).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func UpdateTask(c *gin.Context) {
	wid64, _ := strconv.ParseUint(c.Param("wid"), 10, 64)
	tid64, _ := strconv.ParseUint(c.Param("tid"), 10, 64)

	workspaceID := uint(wid64)
	taskID := uint(tid64)
	userID := c.MustGet("currentUser").(uint)

	var task models.Task
	if err := models.DB.Joins("JOIN projects ON projects.id = tasks.project_id").
		Where("tasks.id = ? AND projects.workspace_id = ?", taskID, workspaceID).
		First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var input struct {
		Status     *string `json:"status"`
		AssigneeID *uint   `json:"assignee_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, role := requireWorkspaceMember(workspaceID, userID)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}

	if input.AssigneeID != nil {
		if role != "OWNER" && role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admin/owner can assign"})
			return
		}
		task.AssigneeID = input.AssigneeID
	}

	if input.Status != nil {
		if role == "MEMBER" && (task.AssigneeID == nil || *task.AssigneeID != userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only assignee or admin can update status"})
			return
		}
		task.Status = *input.Status
	}

	models.DB.Save(&task)
	utils.CreateAuditLog(userID, workspaceID, &task.ProjectID, &task.ID,
		"TASK_UPDATED",
		"Task updated (status/assignee changed)",
	)

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func DeleteTask(c *gin.Context) {
	wid64, _ := strconv.ParseUint(c.Param("wid"), 10, 64)
	tid64, _ := strconv.ParseUint(c.Param("tid"), 10, 64)

	workspaceID := uint(wid64)
	taskID := uint(tid64)
	userID := c.MustGet("currentUser").(uint)

	ok, _ := requireWorkspaceMember(workspaceID, userID, "OWNER", "ADMIN")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
		return
	}

	if err := models.DB.Delete(&models.Task{}, taskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	utils.CreateAuditLog(userID, workspaceID, nil, &taskID,
		"TASK_DELETED",
		"Task deleted",
	)

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
