package routes

import (
	"net/http"
	"os"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/handlers"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/middleware"
	"github.com/gin-gonic/gin"
)

func getUploadDir() string {
	dir := os.Getenv("UPLOAD_DIR")
	if dir == "" {
		dir = "./uploads"
	}
	return dir
}

func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"messages": "Hello, World!",
		})
	})

	api := r.Group("/api")
	{
		// Serve static files for uploads under /api/uploads
		api.Static("/uploads", getUploadDir())

		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
		api.POST("/logout", handlers.Logout)
		api.GET("/me", middleware.AuthRequired(), handlers.Profile)
	}

	user := api.Group("/user")
	{
		user.GET("/", middleware.AuthRequired(), handlers.GetPaginateUser)
		user.GET("/:id", middleware.AuthRequired(), handlers.GetUserByID)
		user.GET("/:id/public", handlers.GetPublicProfile)
		user.POST("/", middleware.AuthRequired(), handlers.CreateUser)
		user.PUT("/:id", middleware.AuthRequired(), handlers.UpdateUser)
		user.PUT("/:id/password", middleware.AuthRequired(), handlers.ChangeUserPassword)
		user.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteUser)
	}

	roles := api.Group("/role")
	{
		roles.GET("/", middleware.AuthRequired(), handlers.GetRoles)
	}

	product := api.Group("/product")
	{
		product.GET("/", middleware.AuthRequired(), handlers.ProductPaginate)
		product.POST("/", middleware.AuthRequired(), handlers.CreateProduct)
		product.PUT("/:id", middleware.AuthRequired(), handlers.UpdateProduct)
		product.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteProduct)

		product.GET("/:id", handlers.ProductDetails)
		product.GET("/self", middleware.AuthRequired(), handlers.GetMyProductsPaginate)
	}

	category := api.Group("/category")
	{
		category.GET("/", middleware.AuthRequired(), handlers.CategoryPaginate)
		category.GET("/all", handlers.GetAllCategory)
		category.POST("/", middleware.AuthRequired(), handlers.CreateCategory)
		category.PUT("/:id", middleware.AuthRequired(), handlers.UpdateCategory)
		category.PUT("/:id/status", middleware.AuthRequired(), handlers.SetCategoryStatus)
		category.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteCategory)
	}

	productCondition := api.Group("/product-condition")
	{
		productCondition.GET("/", handlers.GetAllProductCondition)
		productCondition.POST("/", middleware.AuthRequired(), handlers.CreateProductCondition)
		productCondition.PUT("/:id", middleware.AuthRequired(), handlers.UpdateProductCondition)
		productCondition.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteProductCondition)
	}

	productStatus := api.Group("/product-status")
	{
		productStatus.GET("/", middleware.AuthRequired(), handlers.GetAllProductStatus)
		productStatus.POST("/", middleware.AuthRequired(), handlers.CreateProductStatus)
		productStatus.PUT("/:id", middleware.AuthRequired(), handlers.UpdateProductStatus)
		productStatus.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteProductStatus)
	}

	categoryGroup := api.Group("/category-group")
	{
		categoryGroup.GET("/", middleware.AuthRequired(), handlers.GetAllCategoryGroups)
		categoryGroup.POST("/", middleware.AuthRequired(), handlers.CreateCategoryGroup)
		categoryGroup.PUT("/:id", middleware.AuthRequired(), handlers.UpdateCategoryGroup)
		categoryGroup.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteCategoryGroup)
	}

	categoryStatus := api.Group("/category-status")
	{
		categoryStatus.GET("/", middleware.AuthRequired(), handlers.GetAllCategoryStatuses)
		categoryStatus.POST("/", middleware.AuthRequired(), handlers.CreateCategoryStatus)
		categoryStatus.PUT("/:id", middleware.AuthRequired(), handlers.UpdateCategoryStatus)
		categoryStatus.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteCategoryStatus)
	}

	vehicle := api.Group("/vehicle")
	{
		vehicle.GET("/", middleware.AuthRequired(), handlers.VehiclePaginate)
		vehicle.POST("/", middleware.AuthRequired(), handlers.CreateVehicle)
		vehicle.PUT("/:id", middleware.AuthRequired(), handlers.UpdateVehicle)
		vehicle.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteVehicle)
		vehicle.GET("/:id", handlers.VehicleDetails)
		vehicle.GET("/self", middleware.AuthRequired(), handlers.GetMyVehiclesPaginate)
	}

	vehicleType := api.Group("/vehicle-type")
	{
		vehicleType.GET("/", middleware.AuthRequired(), handlers.GetAllVehicleTypes)
		vehicleType.POST("/", middleware.AuthRequired(), handlers.CreateVehicleType)
		vehicleType.PUT("/:id", middleware.AuthRequired(), handlers.UpdateVehicleType)
		vehicleType.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteVehicleType)
	}

	notification := api.Group("/notification")
	{
		notification.GET("/", middleware.AuthRequired(), handlers.GetNotifications)
		notification.PUT("/:id/read", middleware.AuthRequired(), handlers.MarkNotificationAsRead)
		notification.PUT("/read-all", middleware.AuthRequired(), handlers.MarkAllNotificationsAsRead)
	}

	boost := api.Group("/boost")
	{
		boost.GET("/pricing", handlers.GetBoostPricing)
		boost.POST("/", middleware.AuthRequired(), handlers.CreateBoost)
		boost.GET("/", middleware.AuthRequired(), handlers.GetUserBoosts)
		boost.GET("/:id", middleware.AuthRequired(), handlers.GetBoost)
		boost.PUT("/:id", middleware.AuthRequired(), handlers.UpdateBoost)
		boost.PUT("/:id/cancel", middleware.AuthRequired(), handlers.CancelBoost)
	}

	messages := api.Group("/message")
	{
		messages.GET("/all", middleware.AuthRequired(), handlers.GetAllConversations)
		messages.GET("/conversations", middleware.AuthRequired(), handlers.GetConversations)
		messages.GET("/conversation/:id", middleware.AuthRequired(), handlers.GetMessages)
		messages.POST("/send", middleware.AuthRequired(), handlers.SendMessage)
	}

	reviews := api.Group("/review")
	{
		reviews.GET("/user/:id", handlers.GetUserReviews)
		reviews.POST("/", middleware.AuthRequired(), handlers.CreateReview)
	}

	identity := api.Group("/identity")
	{
		identity.POST("/verify", middleware.AuthRequired(), handlers.VerifyIdentity)
		identity.PUT("/phone", middleware.AuthRequired(), handlers.UpdatePhone)
	}

	api.GET("/listings", handlers.GetAllListings)
	api.GET("/listings/:id", handlers.ListingDetails)
	api.GET("/listings/self", middleware.AuthRequired(), handlers.GetMyListings)
}
