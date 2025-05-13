package s3helper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

var s3Client *s3.Client
var bucketName string
var region string

// Inicializa o S3 client (chamar uma vez no início do app)
func InitS3Helper() error {
	// Tenta carregar .env (ignora erro se não existir, comum em produção)
	_ = godotenv.Load()

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
		log.Println("⚠️  ENVIRONMENT não definido, assumindo 'development'")
	} else {
		log.Println("🌍 Ambiente ativo:", env)
	}

	// Lê variáveis de ambiente
	bucket := os.Getenv("AWS_BUCKET_NAME")
	region = os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if bucket == "" || region == "" || accessKey == "" || secretKey == "" {
		return fmt.Errorf("❌ Variáveis AWS estão faltando (verifique .env ou variáveis do ambiente)")
	}

	bucketName = bucket

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("❌ Erro ao carregar configuração AWS: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)
	log.Println("✅ Cliente S3 configurado com sucesso")

	return nil
}

// Faz upload do arquivo e retorna a URL pública
func UploadFileToS3(file multipart.File, header *multipart.FileHeader, key string) (string, error) {
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
		//ACL:    "public-read", // ou "private"
	})
	if err != nil {
		return "", fmt.Errorf("❌ Falha ao fazer upload para o S3: %v", err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, key)
	return url, nil
}

func DeleteFileFromS3(key string) error {
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	return err
}
