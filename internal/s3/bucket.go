package s3

import "context"

type S3Bucket interface {
	Put(ctx context.Context, data []byte, contentType string, extension string) (string, error)
	Delete(ctx context.Context, key string)
}
