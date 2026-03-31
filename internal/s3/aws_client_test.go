package s3

import (
	"context"
	"testing"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

func TestCreateNewBucket(t *testing.T) {
	const expectedBucket = "my-test-bucket"
	const expectedPrefix = "uploads/"
	keyGen := func() (string, error) { return "unique-key", nil }

	mockS3Client := &awsS3.Client{}
	s := &AWSClient{client: mockS3Client}

	bucket := s.NewBucket(expectedBucket, expectedPrefix, keyGen)

	require.NotNil(t, bucket, "bucket must not be nil")
}

func TestNewAWSClient(t *testing.T) {
	const region = "ru-msk"
	const endpoint = "http://localhost"
	const accessKey = "access-key"
	const secretKey = "secret-key"

	client, err := NewAWSClient(context.TODO(), region, endpoint, accessKey, secretKey)

	require.NoError(t, err, "NewAWSClient must not return error")
	require.NotNil(t, client, "client must not be nil")
	require.NotNil(t, client.client, "aws client must not be nil")
}
