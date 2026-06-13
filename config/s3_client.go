package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client builds an S3-compatible client (AWS S3, Cloudflare R2, MinIO).
func NewS3Client(disk DiskConfig) (*s3.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
		if disk.Endpoint != "" {
			return aws.Endpoint{
				URL:               disk.Endpoint,
				SigningRegion:     disk.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(disk.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(disk.Key, disk.Secret, "")),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = disk.PathStyle
	}), nil
}
