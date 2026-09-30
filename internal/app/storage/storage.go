package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"mime/multipart"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MediaStorage struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewFromEnv() (*MediaStorage, error) {
	endpoint := envOrDefault("MINIO_ENDPOINT", "localhost:9000")
	accessKey := envOrDefault("MINIO_ACCESS_KEY", "root")
	secretKey := envOrDefault("MINIO_SECRET_KEY", "rootpassword")
	bucket := envOrDefault("MINIO_BUCKET", "media")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	publicURL := strings.TrimRight(envOrDefault("MINIO_PUBLIC_URL", scheme+"://"+endpoint), "/")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MediaStorage{client: client, bucket: bucket, publicURL: publicURL}, nil
}

func (s *MediaStorage) Upload(ctx context.Context, file *multipart.FileHeader, folder, contentType, ext string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	name, err := randomName()
	if err != nil {
		return "", err
	}
	objectName := folder + "/" + name + ext

	_, err = s.client.PutObject(ctx, s.bucket, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return s.publicURL + "/" + s.bucket + "/" + objectName, nil
}

func randomName() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
