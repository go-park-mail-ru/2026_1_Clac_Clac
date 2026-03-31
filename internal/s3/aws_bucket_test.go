package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	mockAWSClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3/mock_aws_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAWSBucketPut(t *testing.T) {
	t.Run("successful put", func(t *testing.T) {
		ctx := context.Background()
		mockS3 := new(mockAWSClient.AWSClientAPI)

		b := &AWSBucket{
			client: mockS3,
			bucket: "my-bucket",
			prefix: "test-prefix",
			keyGenerator: func() (string, error) {
				return "file-123", nil
			},
		}

		expectedKey := "test-prefix/file-123.txt"

		mockS3.On("PutObject", ctx, mock.MatchedBy(func(p *awsS3.PutObjectInput) bool {
			return *p.Key == expectedKey && *p.Bucket == "my-bucket"
		}), mock.Anything).Return(&awsS3.PutObjectOutput{}, nil)

		key, err := b.Put(ctx, strings.NewReader("data"), "text/plain", ".txt")

		assert.NoError(t, err, "must not return error")
		assert.Equal(t, expectedKey, key, "keys must be equal")

		mockS3.AssertExpectations(t)
	})

	t.Run("error put", func(t *testing.T) {
		ctx := context.Background()
		mockS3 := new(mockAWSClient.AWSClientAPI)

		b := &AWSBucket{
			client: mockS3,
			bucket: "my-bucket",
			prefix: "test-prefix",
			keyGenerator: func() (string, error) {
				return "file-123", nil
			},
		}

		expectedKey := "test-prefix/file-123.txt"

		mockS3.On("PutObject", ctx, mock.MatchedBy(func(p *awsS3.PutObjectInput) bool {
			return *p.Key == expectedKey && *p.Bucket == "my-bucket"
		}), mock.Anything).Return(nil, errors.New("cannot put file"))

		_, err := b.Put(ctx, strings.NewReader("data"), "text/plain", ".txt")

		assert.Error(t, err, "must return error")
		assert.Contains(t, err.Error(), "cannot put file")

		mockS3.AssertExpectations(t)
	})
}

func TestAWSBucketDelete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		ctx := context.Background()
		bucketName := "my-test-bucket"
		targetKey := "uploads/images/photo.jpg"

		mockS3 := new(mockAWSClient.AWSClientAPI)

		b := &AWSBucket{
			client: mockS3,
			bucket: bucketName,
		}

		mockS3.On("DeleteObject", ctx, mock.MatchedBy(func(input *awsS3.DeleteObjectInput) bool {
			return *input.Bucket == bucketName && *input.Key == targetKey
		}), mock.Anything).Return(&awsS3.DeleteObjectOutput{}, nil)

		err := b.Delete(ctx, targetKey)

		assert.NoError(t, err, "must not return error")

		mockS3.AssertExpectations(t)
	})

	t.Run("error delete", func(t *testing.T) {
		ctx := context.Background()
		bucketName := "my-test-bucket"
		targetKey := "uploads/images/photo.jpg"

		mockS3 := new(mockAWSClient.AWSClientAPI)

		b := &AWSBucket{
			client: mockS3,
			bucket: bucketName,
		}

		mockS3.On("DeleteObject", mock.Anything, mock.Anything, mock.Anything).
			Return((*awsS3.DeleteObjectOutput)(nil), fmt.Errorf("cannot delete file"))

		err := b.Delete(ctx, targetKey)

		assert.Error(t, err, "must return error")
		assert.Contains(t, err.Error(), "cannot delete file")

		mockS3.AssertExpectations(t)
	})
}
