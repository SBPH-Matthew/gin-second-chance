package routes

import (
	"net/http"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/handlers"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"messages": "Hello, World!",
		})
	})

	api := r.Group("/api")
	{
		api.GET("/users", handlers.GetUsers)
		api.POST("/user/create", handlers.CreateUser)
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
		api.GET("/me", middleware.AuthRequired(), handlers.Profile)
	}

	product := api.Group("/product")
	{
		product.GET("/", middleware.AuthRequired(), handlers.ProductPaginate)
		product.POST("/", middleware.AuthRequired(), handlers.CreateProduct)
		product.PUT("/:id", middleware.AuthRequired(), handlers.UpdateProduct)
		product.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteProduct)
	}

}
