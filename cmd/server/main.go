package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"user-service/internal/database"
	"user-service/internal/s3helper"
	"user-service/internal/user"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	if err := s3helper.InitS3Helper(); err != nil {
		log.Fatal("Erro ao iniciar o S3:", err)
	}

	database.ConnectDatabase()

	user.RegisterRoutes(r)

	r.Run(":8087")
}
