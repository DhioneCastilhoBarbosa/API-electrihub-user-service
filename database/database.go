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
	// Tenta carregar .env (ignora erro se não existir, comum em produção)
	_ = godotenv.Load()

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development" // fallback se não for definido
		log.Println("⚠️  ENVIRONMENT não definido, assumindo 'development'")
	} else {
		log.Println("🌍 Ambiente ativo:", env)
	}

	var dsn string
	switch env {
	case "test":
		dsn = os.Getenv("DATABASE_URL_TEST")
	case "production":
		dsn = os.Getenv("DATABASE_URL_PRODUCTION")
	default: // development
		dsn = os.Getenv("DATABASE_URL_DEVELOPMENT")
	}

	if dsn == "" {
		log.Fatalf("❌ DATABASE_URL não encontrada para o ambiente '%s'", env)
	}

	// Conecta ao PostgreSQL
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Falha ao conectar no banco de dados:", err)
	}

	// Migração
	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("❌ Falha ao migrar modelo User:", err)
	}

	log.Println("✅ Banco de dados conectado com sucesso")
}
