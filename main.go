package main

import (
	"net/http"
	"time"
	"user-service/database"
	"user-service/middlewares"
	"user-service/models"
	"user-service/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"username"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func main() {
	r := gin.Default()

	// Configuração do CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Em produção, substitua por domínios confiáveis
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	database.ConnectDatabase()

	r.POST("/user/register", func(c *gin.Context) {
		var newUser models.User
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if err := newUser.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		result := database.DB.Create(&newUser)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	})

	r.POST("/user/login", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		var user models.User
		if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		if !user.CheckPassword(input.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}

		token, err := utils.GenerateJWT(user.ID, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	r.GET("/user/list", middlewares.AuthMiddleware(), func(c *gin.Context) {
		var users []models.User
		database.DB.Find(&users)
		var userResponses []UserResponse
		for _, user := range users {
			userResponses = append(userResponses, UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			})
		}
		c.JSON(http.StatusOK, userResponses)
	})

	r.Run(":8082")
}
