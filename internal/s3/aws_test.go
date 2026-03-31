package s3_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awsTypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUploadFile(t *testing.T) {
	const region = "ru-msk"
	const endpoint = "https://hb.ru-msk.vkcloud-storage.ru"
	const bucketName = "nexus-avatars-prod"
	const accessKey = ""
	const privateKey = ""

	client, err := s3.NewAWSClient(context.Background(), region, endpoint, accessKey, privateKey)
	require.NoError(t, err, "must not return error")

	bucket := client.NewBucket(bucketName, "avatar", func() (string, error) {
		return uuid.New().String(), nil
	}, awsTypes.ObjectCannedACLPublicRead)

	file, err := os.Open("image.webp")
	require.NoError(t, err, "must not return error")

	defer file.Close()

	objectKey, err := bucket.Put(context.Background(), file, "image/webp", ".webp")
	require.NoError(t, err, "must not return error")
	fmt.Println(objectKey)
}
