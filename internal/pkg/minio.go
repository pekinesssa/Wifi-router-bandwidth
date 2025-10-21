package pkg

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	// "github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)


type Client struct{
	client *minio.Client
	bucketName string
}

func NewClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*Client, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing minio configuration")
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", bucketName)
	}

	return &Client{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

func FromEnv() (*Client, error){
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	return NewClient(endpoint, accessKeyID, secretAccessKey, bucketName, useSSL)
}

func (c *Client) DeleteObject(ctx context.Context, objName string) error {
	options := minio.RemoveObjectOptions{}
	return c.client.RemoveObject(ctx, c.bucketName, objName, options)
}

func (c *Client) UploadObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := c.client.PutObject(
		ctx,
		c.bucketName,
		objectName,
		reader,
		objectSize,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func (c *Client) GetBucketName() string {
    return c.bucketName
}
