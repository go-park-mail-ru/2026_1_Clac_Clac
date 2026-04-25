package s3

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog"
)

//go:generate mockery --name AWSClientAPI --output mock_aws_client
type AWSClientAPI interface {
	PutObject(ctx context.Context, params *awsS3.PutObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *awsS3.DeleteObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteObjectOutput, error)
}

type AWSBucket struct {
	client AWSClientAPI
	bucket string
	prefix string
	acl    awsTypes.ObjectCannedACL
}

// Подставляет префикс
func (b *AWSBucket) Put(ctx context.Context, data io.Reader, key string, contentType string) (string, error) {
	objectKey := path.Join(b.prefix, key)

	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("bucket", b.bucket).
		Str("key", objectKey).
		Str("content_type", contentType).
		Str("acl", string(b.acl)).
		Msg("s3 PutObject request")

	_, err := b.client.PutObject(ctx, &awsS3.PutObjectInput{
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
