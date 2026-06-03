package services

import (
	"be-lms/config"
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadToS3(fileData []byte, filename string) (string, error) {
	s3Disk := config.Disks[config.S3]

	// Create S3 client
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
		if s3Disk.Endpoint != "" {
			ep := aws.Endpoint{
				URL:               s3Disk.Endpoint,
				SigningRegion:     s3Disk.Region,
				HostnameImmutable: true,
			}
			return ep, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(s3Disk.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Disk.Key, s3Disk.Secret, "")),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s3Disk.PathStyle
	})

	// Upload to S3 in exports folder
	key := fmt.Sprintf("public/exports/%s", filename)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Disk.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileData),
		ContentType: aws.String("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// Return public URL
	objectURL := fmt.Sprintf("%s/%s", strings.TrimRight(s3Disk.URL, "/"), key)
	return objectURL, nil
}
