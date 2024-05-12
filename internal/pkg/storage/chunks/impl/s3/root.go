package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(client *minio.Client) *MinioStorage {
	return &MinioStorage{client: client}
}

func (s *MinioStorage) CheckOrCreateBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("s.client.BucketExists: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("s.client.MakeBucket: %w", err)
		}
	}

	return nil
}

func (s *MinioStorage) GetObjectDownloadLink(ctx context.Context, bucketName, objectName string) (*url.URL, error) {
	return s.client.PresignedGetObject(ctx, bucketName, objectName, time.Hour*24*7, nil)
}

func (s *MinioStorage) UploadObjectMultipart(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) error {
	err := s.CheckOrCreateBucket(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("s.CheckOrCreateBucket: %w", err)
	}

	_, err = s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: "text/csv",
	})
	if err != nil {
		return fmt.Errorf("s.client.PutObject: %w", err)
	}

	return nil
}

func (s *MinioStorage) GetObjectBytes(ctx context.Context, bucketName, objectName string, from, to int64) (*minio.Object, error) {
	opts := minio.GetObjectOptions{}
	err := opts.SetRange(from, to)
	if err != nil {
		return nil, fmt.Errorf("opts.SetRange: %w", err)
	}

	object, err := s.client.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, fmt.Errorf("s.client.GetObject: %w", err)
	}

	return object, nil
}
