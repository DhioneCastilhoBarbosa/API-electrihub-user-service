package s3helper

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	S3Client *s3.Client
	Bucket   = "eletrihub-users-files"
	Region   = "us-east-2" // ex: "us-east-1"
)

func InitS3() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(Region))
	if err != nil {
		return err
	}
	S3Client = s3.NewFromConfig(cfg)
	return nil
}

func UploadFileToS3(file multipart.File, fileHeader *multipart.FileHeader, key string) (string, error) {
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return "", err
	}

	_, err = S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    "public-read", // opcional: se quiser acesso público
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", Bucket, Region, key)
	return url, nil
}
