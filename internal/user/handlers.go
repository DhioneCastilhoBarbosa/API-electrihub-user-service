package user

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"user-service/internal/database"
	"user-service/internal/user/models"

	"user-service/internal/s3helper"
	"user-service/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID           string `json:"id"`
	Name         string `json:"username"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Phone        string `json:"phone"`
	CPF          string `json:"cpf"`
	CNPJ         string `json:"cnpj"`
	CompanyName  string `json:"company_name"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Complement   string `json:"complement"`
	CEP          string `json:"cep"`
	BirthDate    string `json:"birth_date"`
	Reference    string `json:"reference"`
	AceptTerms   bool   `json:"accept_terms"`
	Photo        string `json:"photo"`
}

func RegisterUser(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

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
}

func LoginUser(c *gin.Context) {
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
}

func ListUsers(c *gin.Context) {
	role := c.Query("role")
	id := c.Query("id")
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
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email,
			Phone:        user.Phone,
			CPF:          user.CPF,
			CNPJ:         user.CNPJ,
			CompanyName:  user.CompanyName,
			Street:       user.Street,
			Number:       user.Number,
			Neighborhood: user.Neighborhood,
			City:         user.City,
			State:        user.State,
			Complement:   user.Complement,
			CEP:          user.CEP,
			BirthDate:    user.BirthDate,
			Reference:    user.Reference,
			AceptTerms:   user.AceptTerms,
			Role:         user.Role,
			Photo:        user.Photo,
		})
	}

	c.JSON(http.StatusOK, userResponses)
}

func ListPendingInstallers(c *gin.Context) {
	var users []models.User
	database.DB.Where("role = ? AND authorized = ?", "instalador", false).Find(&users)
	c.JSON(http.StatusOK, users)
}

func AuthorizeUser(c *gin.Context) {
	id := c.Param("id")

	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Update("authorized", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao autorizar usuário"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário autorizado com sucesso"})
}

func UpdatePassword(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		NewPassword string `json:"new_password"`
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
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	delete(updateData, "id")
	delete(updateData, "email")
	delete(updateData, "password")
	delete(updateData, "role")
	delete(updateData, "authorized")

	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
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
}

func UpdateUserPhoto(c *gin.Context) {
	id := c.Param("id")

	// Lê o arquivo da requisição
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Imagem não enviada"})
		return
	}

	// Busca usuário no banco
	var user models.User
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Se já houver uma foto, deleta ela do S3
	if user.Photo != "" {
		// Exemplo: https://bucket.s3.region.amazonaws.com/users/ID_TIMESTAMP.png
		parts := strings.Split(user.Photo, "/")
		if len(parts) > 0 {
			oldKey := strings.Join(parts[len(parts)-2:], "/") // pega os dois últimos: "users/arquivo.png"
			if err := s3helper.DeleteFileFromS3(oldKey); err != nil {
				fmt.Println("⚠️ Erro ao deletar arquivo anterior:", err)
			}
		}
	}

	// Abre o novo arquivo
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir imagem"})
		return
	}

	// Gera o novo nome
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("users/%s_%d%s", id, time.Now().Unix(), ext)

	// Faz upload para S3
	url, err := s3helper.UploadFileToS3(src, file, fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao fazer upload para S3"})
		return
	}

	// Atualiza no banco
	if err := database.DB.Model(&user).Update("photo", url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar URL da imagem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto atualizada com sucesso",
		"photo":   url,
	})
}
