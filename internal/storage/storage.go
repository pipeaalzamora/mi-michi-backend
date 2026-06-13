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

// UploadCatPhoto sube una foto al bucket S3 y devuelve la key privada del objeto.
func UploadCatPhoto(ctx context.Context, userID, catID, filename string, file io.Reader) (string, error) {
	bucket, _, err := s3Config()
	if err != nil {
		return "", err
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

	client, err := newS3Client(ctx)
	if err != nil {
		return "", err
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo a S3: %w", err)
	}

	return key, nil
}

func PresignCatPhoto(ctx context.Context, key string, expires time.Duration) (string, error) {
	bucket, _, err := s3Config()
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", fmt.Errorf("key de foto requerida")
	}
	if expires <= 0 {
		expires = 6 * time.Hour
	}

	client, err := newS3Client(ctx)
	if err != nil {
		return "", err
	}

	presigner := s3.NewPresignClient(client)
	result, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("error firmando foto S3: %w", err)
	}
	return result.URL, nil
}

func DeleteCatPhoto(ctx context.Context, key string) error {
	bucket, _, err := s3Config()
	if err != nil {
		return err
	}
	if key == "" {
		return nil
	}

	client, err := newS3Client(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error eliminando foto S3: %w", err)
	}
	return nil
}

func newS3Client(ctx context.Context) (*s3.Client, error) {
	_, region, err := s3Config()
	if err != nil {
		return nil, err
	}

	optFns := []func(*config.LoadOptions) error{config.WithRegion(region)}
	accessKey := firstEnv("S3_ACCESS_KEY", "AWS_ACCESS_KEY_ID")
	secretKey := firstEnv("S3_SECRET_KEY", "AWS_SECRET_ACCESS_KEY")
	sessionToken := firstEnv("S3_SESSION_TOKEN", "AWS_SESSION_TOKEN")
	if accessKey != "" && secretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("error configurando AWS: %w", err)
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	clientOpts := []func(*s3.Options){}
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	return s3.NewFromConfig(cfg, clientOpts...), nil
}

func s3Config() (bucket, region string, err error) {
	bucket = os.Getenv("S3_BUCKET")
	region = os.Getenv("S3_REGION")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if bucket == "" || region == "" {
		return "", "", fmt.Errorf("S3_BUCKET y S3_REGION son requeridos")
	}
	return bucket, region, nil
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
