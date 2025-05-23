package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                    string    `json:"id" gorm:"type:text;primaryKey"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email" gorm:"unique"`
	Password              string    `json:"password"`
	Phone                 string    `json:"phone"`
	CPF                   string    `json:"cpf"`
	CNPJ                  string    `json:"cnpj"`
	CompanyName           string    `json:"company_name"`
	Street                string    `json:"street"`
	Number                string    `json:"number"`
	Neighborhood          string    `json:"neighborhood"`
	City                  string    `json:"city"`
	State                 string    `json:"state"`
	Complement            string    `json:"complement"`
	CEP                   string    `json:"cep"`
	Latitude              float64   `json:"latitude"`
	Longitude             float64   `json:"longitude"`
	BirthDate             string    `json:"birth_date"`
	Reference             string    `json:"reference"`
	AceptTerms            bool      `json:"accept_terms"`
	Role                  string    `json:"role"`
	Authorized            bool      `json:"authorized" gorm:"default:false"`
	AverageRating         float64   `json:"average_rating"`
	TotalServicesAccepted int       `json:"total_services_accepted"`
	ServicesNotExecuted   int       `json:"services_not_executed"`
	Photo                 string    `json:"photo"`
	CreatedAt             time.Time `json:"-"`
	UpdatedAt             time.Time `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
