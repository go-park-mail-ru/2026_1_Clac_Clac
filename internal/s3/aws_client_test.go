package s3

import (
	"context"
	"testing"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNewBucket(t *testing.T) {
	const expectedBucket = "my-test-bucket"
	const expectedPrefix = "uploads/"
	const expectedACL = awsTypes.ObjectCannedACLPrivate
	keyGen := func() (string, error) { return "unique-key", nil }

	mockS3Client := &awsS3.Client{}
	s := &AWSClient{client: mockS3Client}

	bucket := s.NewBucket(expectedBucket, expectedPrefix, keyGen, expectedACL)

	require.NotNil(t, bucket, "bucket must not be nil")
	assert.Equal(t, mockS3Client, bucket.client, "aws s3 clients must be equal")
	assert.Equal(t, expectedBucket, bucket.bucket, "bucket names must be equal")
	assert.Equal(t, expectedPrefix, bucket.prefix, "prefixes must be equal")
	assert.Equal(t, expectedACL, bucket.acl, "acl must be equal")

	key, err := bucket.keyGenerator()
	assert.NoError(t, err, "key generator must no return error")
	assert.Equal(t, "unique-key", key, "keys must be equal")
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
