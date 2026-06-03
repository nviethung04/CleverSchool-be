package jobs

import (
	"be-lms/config"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/robfig/cron/v3"
)

func StartS3CleanupCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// Run every 5 minutes (for testing). Replace with "0 4 * * *" for daily 4AM run.
	c.AddFunc("0 4 * * *", func() {
		retentionDays := getEnvInt("S3_RETENTION_DAYS", 7)
		cutoff := time.Now().AddDate(0, 0, -retentionDays)

		paths := []string{"public/exports/", "logs/"}
		for _, path := range paths {
			config.Log.Infof("Starting S3 cleanup for prefix '%s' - removing files older than %s", path, cutoff.Format("2006-01-02"))

			_, err := deleteOldS3Objects(path, cutoff)
			if err != nil {
				config.Log.Errorf("Failed to clean up S3 path '%s': %v", path, err)
				continue
			}

			config.Log.Infof("✅ S3 cleanup for '%s': deleted objects older than %s", path, cutoff.Format("2006-01-02"))
		}
	})

	c.Start()
}

func deleteOldS3Objects(prefix string, beforeTime time.Time) (int, error) {
	s3Disk := config.Disks[config.S3]

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if s3Disk.Endpoint != "" {
			return aws.Endpoint{
				URL:               s3Disk.Endpoint,
				SigningRegion:     s3Disk.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(s3Disk.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Disk.Key, s3Disk.Secret, "")),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s3Disk.PathStyle
	})

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Disk.Bucket),
		Prefix: aws.String(strings.TrimLeft(prefix, "/")),
	})

	totalDeleted := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return totalDeleted, fmt.Errorf("ListObjectsV2 failed: %w", err)
		}

		var candidates []s3types.ObjectIdentifier
		for _, obj := range page.Contents {
			if obj.LastModified != nil && obj.LastModified.Before(beforeTime) &&
				obj.StorageClass != s3types.ObjectStorageClassGlacier &&
				obj.StorageClass != s3types.ObjectStorageClassDeepArchive {

				candidates = append(candidates, s3types.ObjectIdentifier{Key: obj.Key})
			}
		}

		// Delete in batches of 1000 (S3 API limit)
		for len(candidates) > 0 {
			batch := candidates
			if len(batch) > 1000 {
				batch = candidates[:1000]
				candidates = candidates[1000:]
			} else {
				candidates = nil
			}

			resp, err := client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
				Bucket: aws.String(s3Disk.Bucket),
				Delete: &s3types.Delete{
					Objects: batch,
					Quiet:   aws.Bool(true),
				},
			})
			if err != nil {
				config.Log.Errorf("DeleteObjects failed: %v", err)
				continue
			}

			totalDeleted += len(resp.Deleted)
		}
	}

	return totalDeleted, nil
}

func getEnvInt(key string, defaultValue int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultValue
	}
	return n
}
