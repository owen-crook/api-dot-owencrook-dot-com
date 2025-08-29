package gcs

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Client struct {
	client *storage.Client
}

func NewGCSClient(ctx context.Context) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

func (g *Client) UploadFile(ctx context.Context, bucketName, objectPath string, data io.Reader, contentType string) error {
	bucket := g.client.Bucket(bucketName)
	object := bucket.Object(objectPath)

	writer := object.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(writer, data); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

func (g *Client) DeleteFile(ctx context.Context, bucketName, objectPath string) error {
	bucket := g.client.Bucket(bucketName)
	object := bucket.Object(objectPath)

	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete GCS object: %w", err)
	}

	return nil
}
