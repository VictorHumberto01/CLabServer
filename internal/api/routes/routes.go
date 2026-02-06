package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitub/CLabServer/internal/api/handlers"
	"github.com/vitub/CLabServer/internal/api/middleware"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/compile", middleware.OptionalAuth, handlers.HandleCompile)

	r.OPTIONS("/compile", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/login", handlers.LoginWithToken)
	r.POST("/login/cookie", handlers.LoginWithCookie)
	r.POST("/login/matricula", handlers.LoginMatricula)
	r.GET("/validate", middleware.RequireAuth, handlers.Validate)
	r.PUT("/profile", middleware.RequireAuth, handlers.UpdateProfile)

	classrooms := r.Group("/classrooms")
	classrooms.Use(middleware.RequireAuth)
	{
		classrooms.POST("", handlers.CreateClassroom)
		classrooms.GET("", handlers.ListClassrooms)
		classrooms.POST("/:id/students", handlers.AddStudent)
		classrooms.DELETE("/:id/students/:studentId", handlers.RemoveStudent)
		classrooms.DELETE("/:id", handlers.DeleteClassroom)
		classrooms.POST("/:id/topics", handlers.CreateTopic)
		classrooms.GET("/:id/topics", handlers.ListTopics)
		classrooms.POST("/:id/exercises", handlers.CreateExercise)
		classrooms.GET("/:id/exercises", handlers.ListExercises)
		classrooms.POST("/:id/exam", handlers.ToggleExamMode)
	}

	history := r.Group("/history")
	history.Use(middleware.RequireAuth)
	{
		history.GET("", handlers.ListHistory)
	}

	users := r.Group("/users")
	users.Use(middleware.RequireAuth)
	{
		users.POST("", handlers.CreateUser)
		users.GET("", handlers.ListUsers)
	}

}
