package database

import (
	"log"
	"os"
	"user-service/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Carrega variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Aviso: arquivo .env não carregado (usando variáveis de ambiente)")
	}

	env := os.Getenv("ENVIRONMENT")
	var dsn string

	switch env {
	case "test":
		dsn = os.Getenv("DATABASE_URL_TEST")
	case "production":
		dsn = os.Getenv("DATABASE_URL_PRODUCTION")
	default:
		dsn = os.Getenv("DATABASE_URL_DEVELOPMENT")
	}

	if dsn == "" {
		log.Fatal("❌ DATABASE_URL não encontrada para o ambiente: ", env)
	}

	// Conecta ao PostgreSQL
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Falha ao conectar no banco de dados:", err)
	}

	// Migração do modelo
	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("❌ Falha ao migrar modelo User:", err)
	}

	log.Println("✅ Banco de dados conectado com sucesso")
}
