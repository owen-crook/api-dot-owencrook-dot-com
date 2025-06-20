package gcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Client struct {
	client     *storage.Client
	bucketName string
}

func NewGCSClient(ctx context.Context, bucketName string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (g *Client) UploadFile(ctx context.Context, objectPath string, data io.Reader, contentType string) (string, error) {
	bucket := g.client.Bucket(g.bucketName)
	object := bucket.Object(objectPath)

	writer := object.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(writer, data); err != nil {
		_ = writer.Close()
		return "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	// // Optional: make the object public
	// if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
	// 	return "", fmt.Errorf("failed to set public ACL: %w", err)
	// }

	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)
	return publicURL, nil
}
