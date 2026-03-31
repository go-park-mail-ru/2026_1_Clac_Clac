package s3

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWSBucket struct {
	client       *awsS3.Client
	bucket       string
	prefix       string
	acl          awsTypes.ObjectCannedACL
	keyGenerator func() (string, error)
}

// Подставляет префикс
func (b *AWSBucket) Put(ctx context.Context, data io.Reader, contentType string, extension string) (string, error) {
	key, err := b.keyGenerator()
	if err != nil {
		return "", fmt.Errorf("cannot generate key: %w", err)
	}
	objectKey := path.Join(b.prefix, key+strings.ToLower(extension))

	_, err = b.client.PutObject(ctx, &awsS3.PutObjectInput{
		Bucket:      aws.String(b.bucket),
		Key:         aws.String(objectKey),
		Body:        data,
		ContentType: aws.String(contentType),
		ACL:         b.acl,
	})
	if err != nil {
		return "", fmt.Errorf("aws s3 cannot put object: %w", err)
	}

	return objectKey, nil
}

// Не подставляет префикс
func (b *AWSBucket) Delete(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &awsS3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("aws s3 delete object error: %w", err)
	}

	return nil
}
