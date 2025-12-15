package main

import (
	"log"

	_ "github.com/SBPH-Matthew/second-chance/cmd/docs"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading env %v", err)
	}
}

func main() {
	database.Connect()
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server %v", err)
	}
}
