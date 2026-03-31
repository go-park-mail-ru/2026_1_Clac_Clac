package s3

import (
	"context"
	"io"
)

type S3Client interface {
	NewBucket(bucket string, prefix string, keyGenerator func() (string, error)) S3Bucket
}

type S3Bucket interface {
	Put(ctx context.Context, data io.Reader, contentType string, extension string) (string, error)
	Delete(ctx context.Context, key string) error
}
