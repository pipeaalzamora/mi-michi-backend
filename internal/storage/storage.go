package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadCatPhoto sube una foto al bucket S3 y devuelve la URL pública.
func UploadCatPhoto(ctx context.Context, userID, catID, filename string, file io.Reader) (string, error) {
	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	endpoint := os.Getenv("S3_ENDPOINT") // vacío para AWS real, útil para MinIO/R2

	if bucket == "" || region == "" || accessKey == "" || secretKey == "" {
		return "", fmt.Errorf("S3_BUCKET, S3_REGION, S3_ACCESS_KEY y S3_SECRET_KEY son requeridos")
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("cats/%s/%s/%d%s", userID, catID, time.Now().UnixMilli(), ext)

	body, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "image/jpeg"
	}

	// Configurar cliente S3
	optFns := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return "", fmt.Errorf("error configurando AWS: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(cfg, clientOpts...)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo a S3: %w", err)
	}

	// URL pública
	var publicURL string
	if endpoint != "" {
		// MinIO / Cloudflare R2 / custom
		publicURL = fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
	} else {
		// AWS S3 estándar
		publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	}

	return publicURL, nil
}
