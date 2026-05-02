package s3

import (
	"context"
	"io"
)

type Action string

var ACL = struct {
	PublicRead Action
}{
	PublicRead: "acl-public-read",
}

type S3Client interface {
	NewBucket(bucket string, prefix string, action Action) S3Bucket
}

type S3Bucket interface {
	Put(ctx context.Context, data io.Reader, key string, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
}
