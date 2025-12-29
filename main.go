package main

import (
	"multi-tenet/controllers"
	"multi-tenet/middleware"
	"multi-tenet/models"

	"github.com/gin-gonic/gin"
)

func main() {
	models.ConnnectDatabase()

	r := gin.Default()
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	protected := r.Group("/api")

	protected.Use(middleware.JwtHandler())
	{
		protected.POST("/workspaces", controllers.CreateWorkspace)
		protected.GET("/workspaces", controllers.Listworkspace)

		protected.POST("/workspaces/:wid/invite/:uid", controllers.AddNewMember)
		protected.PATCH("/workspaces/:wid/members/:uid/role/:role", controllers.UpdateRole)

		protected.POST("/workspaces/:wid/projects", controllers.CreateProject)
		protected.GET("/workspaces/:wid/projects", controllers.GetProjects)

		protected.POST("/workspaces/:wid/projects/:pid/tasks", controllers.CreateTask)
		protected.GET("/workspaces/:wid/projects/:pid/tasks", controllers.GetTasks)

		protected.PATCH("/workspaces/:wid/tasks/:tid", controllers.UpdateTask)
		protected.DELETE("/workspaces/:wid/tasks/:tid", controllers.DeleteTask)
	}
	r.Run(":8080")
}
