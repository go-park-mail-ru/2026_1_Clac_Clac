package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsCredentials "github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWSClient struct {
	client *awsS3.Client
}

func (s *AWSClient) NewBucket(bucket string, prefix string, action Action) S3Bucket {
	var acl = awsTypes.ObjectCannedACLPrivate

	switch action {
	case ACL.PublicRead:
		acl = awsTypes.ObjectCannedACLPublicRead
	}

	return &AWSBucket{
		client: s.client,
		bucket: bucket,
		prefix: prefix,
		acl:    acl,
	}
}

func NewAWSClient(ctx context.Context, region, endpoint, access_key, secret_key string) (*AWSClient, error) {
	conf, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(region),
		awsConfig.WithCredentialsProvider(awsCredentials.NewStaticCredentialsProvider(
			access_key,
			secret_key,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("error when loading config: %w", err)
	}

	client := awsS3.NewFromConfig(conf, func(o *awsS3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &AWSClient{
		client: client,
	}, nil
}
