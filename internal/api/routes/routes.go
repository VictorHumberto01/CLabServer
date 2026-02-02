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

	r.POST("/compile", handlers.HandleCompile)

	r.OPTIONS("/compile", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/signup", handlers.SignUp)
	r.POST("/login", handlers.LoginWithToken)
	r.POST("/login/cookie", handlers.LoginWithCookie)
	r.GET("/validate", middleware.RequireAuth, handlers.Validate)

	classrooms := r.Group("/classrooms")
	classrooms.Use(middleware.RequireAuth)
	{
		classrooms.POST("", handlers.CreateClassroom)
		classrooms.GET("", handlers.ListClassrooms)
		classrooms.POST("/:id/students", handlers.AddStudent)
	}
}
