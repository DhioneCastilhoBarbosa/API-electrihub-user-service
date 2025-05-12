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
	ID         string `json:"id"`
	Name       string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Phone      string `json:"phone"`
	CPF        string `json:"cpf"`
	Address    string `json:"address"`
	CEP        string `json:"cep"`
	BirthDate  string `json:"birth_date"`
	Reference  string `json:"reference"`
	AceptTerms bool   `json:"accept_terms"`
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	database.ConnectDatabase()

	// Cadastro de usuário
	r.POST("/user/register", func(c *gin.Context) {
		var newUser models.User
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Define autorização com base no papel
		if newUser.Role == "cliente" {
			newUser.Authorized = true
		} else {
			newUser.Authorized = false
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

	// Login
	// Login
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

		// Verifica se o usuário está autorizado
		if !user.Authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Usuário ainda não autorizado"})
			return
		}

		token, err := utils.GenerateJWT(user.ID, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ID":    user.ID,
			"name":  user.Name,
			"token": token,
		})
	})

	// Lista todos os usuários, ou filtra por role e/ou ID
	r.GET("/user/list", middlewares.AuthMiddleware(), func(c *gin.Context) {
		role := c.Query("role")
		id := c.Query("id") // novo filtro por ID
		var users []models.User
		query := database.DB

		if id != "" {
			query = query.Where("id = ?", id)
		}

		if role != "" {
			query = query.Where("role = ?", role)
		}

		if err := query.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar usuários"})
			return
		}

		var userResponses []UserResponse
		for _, user := range users {
			userResponses = append(userResponses, UserResponse{
				ID:         user.ID,
				Name:       user.Name,
				Email:      user.Email,
				Phone:      user.Phone,
				CPF:        user.CPF,
				Address:    user.Address,
				CEP:        user.CEP,
				BirthDate:  user.BirthDate,
				Reference:  user.Reference,
				AceptTerms: user.AceptTerms,
				Role:       user.Role,
			})
		}

		c.JSON(http.StatusOK, userResponses)
	})

	// Lista instaladores pendentes de autorização
	r.GET("/user/installers/pending", middlewares.AuthMiddleware(), func(c *gin.Context) {
		var users []models.User
		database.DB.Where("role = ? AND authorized = ?", "instalador", false).Find(&users)
		c.JSON(http.StatusOK, users)
	})

	// Autoriza usuário (admin)
	r.PATCH("/user/:id/authorize", middlewares.AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")

		if err := database.DB.Model(&models.User{}).Where("id = ?", id).Update("authorized", true).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao autorizar usuário"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Usuário autorizado com sucesso"})
	})

	// Atualiza a senha do usuário
	r.PUT("/user/:id/password", middlewares.AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")

		var body struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		var user models.User
		if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Verifica se a senha atual está correta
		if !user.CheckPassword(body.CurrentPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect current password"})
			return
		}

		// Atualiza a senha
		user.Password = body.NewPassword
		if err := user.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
			return
		}

		if err := database.DB.Model(&user).Update("password", user.Password).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
	})

	// Atualiza os dados de um usuário
	r.PUT("/user/:id", middlewares.AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Impede alterações em campos sensíveis
		delete(updateData, "id")
		delete(updateData, "email")      // evite mudar e-mail diretamente
		delete(updateData, "password")   // use rota separada para troca de senha
		delete(updateData, "role")       // evite que o usuário mude o próprio papel
		delete(updateData, "authorized") // evite autoconcessão de permissão

		result := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updateData)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	})

	// Deleta um usuário pelo ID
	r.DELETE("/user/:id", middlewares.AuthMiddleware(), func(c *gin.Context) {
		id := c.Param("id")

		var user models.User
		if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if err := database.DB.Delete(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	})

	r.Run(":8087")
}
