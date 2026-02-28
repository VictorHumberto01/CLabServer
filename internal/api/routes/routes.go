package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/handlers"
	"github.com/vitub/CLabServer/internal/api/middleware"
	"github.com/vitub/CLabServer/internal/ws"
)

func SetupRoutes(r *gin.Engine, hub *ws.Hub) {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/ws", middleware.OptionalAuth, func(c *gin.Context) {
		ws.ServeWs(hub, c)
	})

	r.POST("/compile", middleware.OptionalAuth, handlers.HandleCompile)
	r.OPTIONS("/compile", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	auth := r.Group("")
	{
		auth.POST("/login", handlers.LoginWithToken)
		auth.POST("/login/cookie", handlers.LoginWithCookie)
		auth.POST("/login/matricula", handlers.LoginMatricula)
		auth.GET("/validate", middleware.RequireAuth, handlers.Validate)
	}

	r.PUT("/profile", middleware.RequireAuth, handlers.UpdateProfile)
	admin := r.Group("/admin")
	{
		admin.POST("/create-teacher", handlers.CreateTeacher)
	}

	users := r.Group("/users")
	users.Use(middleware.RequireAuth)
	{
		users.POST("", handlers.CreateUser)
		users.GET("", handlers.ListUsers)
	}

	classrooms := r.Group("/classrooms")
	classrooms.Use(middleware.RequireAuth)
	{
		classrooms.POST("", handlers.CreateClassroom)
		classrooms.GET("", handlers.ListClassrooms)
		classrooms.DELETE("/:id", handlers.DeleteClassroom)

		classrooms.POST("/:id/teachers", handlers.AddTeacher)
		classrooms.DELETE("/:id/teachers/:teacherId", handlers.RemoveTeacher)

		classrooms.POST("/:id/students", handlers.AddStudent)
		classrooms.DELETE("/:id/students/:studentId", handlers.RemoveStudent)

		classrooms.POST("/:id/topics", handlers.CreateTopic)
		classrooms.GET("/:id/topics", handlers.ListTopics)
		classrooms.DELETE("/:id/topics/:topicId", handlers.DeleteTopic)

		classrooms.POST("/:id/exercises", handlers.CreateExercise)
		classrooms.GET("/:id/exercises", handlers.ListExercises)

		classrooms.POST("/:id/exam", func(c *gin.Context) {
			handlers.ToggleExamMode(c, hub)
		})
	}

	history := r.Group("/history")
	history.Use(middleware.RequireAuth)
	{
		history.GET("", handlers.ListHistory)
	}
}
