package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"be-cleverschool/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type H5PMetadata struct {
	Title   string         `json:"title"`
	Author  string         `json:"author"`
	License string         `json:"license"`
	Content map[string]any `json:"content"`
}

func ExtractH5PMetadata(filePath string, outputDir string) (*H5PMetadata, any, error) {
	if exists, _ := dirHasFiles(outputDir); exists {
		fmt.Println("Skip extraction, already exists:", outputDir)
		return parseH5PJson(outputDir)
	}

	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open file: %v", err)
	}
	defer r.Close()

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, nil, fmt.Errorf("cannot create output directory: %v", err)
	}

	for _, f := range r.File {
		destPath := filepath.Join(outputDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return nil, nil, fmt.Errorf("cannot create directory: %v", err)
			}
			continue
		}

		if err := extractFile(f, destPath); err != nil {
			return nil, nil, fmt.Errorf("cannot extract file: %v", err)
		}
	}

	return parseH5PJson(outputDir)
}

func dirHasFiles(path string) (bool, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return len(files) > 0, nil
}

func parseH5PJson(outputDir string) (*H5PMetadata, any, error) {
	var metadata *H5PMetadata
	var content any

	h5pPath := filepath.Join(outputDir, "h5p.json")
	contentPath := filepath.Join(outputDir, "content", "content.json")

	if data, err := os.ReadFile(h5pPath); err == nil {
		_ = json.Unmarshal(data, &metadata)
	}

	if data, err := os.ReadFile(contentPath); err == nil {
		_ = json.Unmarshal(data, &content)
	}

	if metadata == nil {
		return nil, nil, fmt.Errorf("h5p.json not found")
	}

	return metadata, content, nil
}

func ExtractH5PMetadataFromS3(key string, outputDir string) (*H5PMetadata, any, error) {
	disk := config.Disks["s3"]

	bucket := disk.Bucket
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(disk.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			disk.Key,
			disk.Secret,
			"",
		)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	opts := s3.Options{
		Credentials: awsCfg.Credentials,
		Region:      disk.Region,
	}
	if disk.Endpoint != "" {
		opts.BaseEndpoint = aws.String(disk.Endpoint)
		opts.UsePathStyle = disk.PathStyle
	}
	s3Client := s3.New(opts)

	output, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, output.Body); err != nil {
		return nil, nil, fmt.Errorf("failed to read S3 object body: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open zip from S3: %v", err)
	}

	var metadata *H5PMetadata
	var content any

	var isExtract = true

	for _, f := range r.File {
		if f.Name == "h5p.json" {
			isExtract = false
		}
	}

	if isExtract {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return nil, nil, fmt.Errorf("cannot create output directory: %v", err)
		}
	}

	for _, f := range r.File {
		if isExtract {
			destPath := filepath.Join(outputDir, f.Name)

			if err := extractFile(f, destPath); err != nil {
				return nil, nil, fmt.Errorf("cannot extract file: %v", err)
			}

			err := uploadFileToS3(s3Client, bucket, outputDir + "/"+f.Name, destPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to upload file to S3: %v", err)
			}
		}

		if f.Name == "h5p.json" {
			rc, _ := f.Open()
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			_ = json.Unmarshal(data, &metadata)
		}

		if f.Name == "content/content.json" {
			rc, _ := f.Open()
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			_ = json.Unmarshal(data, &content)
		}
	}

	if metadata == nil {
		return nil, nil, fmt.Errorf("h5p.json not found")
	}

	fmt.Println("Files uploaded to:", outputDir)

	return metadata, content, nil
}

func extractFile(f *zip.File, destPath string) error {
	if _, err := os.Stat(destPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, rc)
	return err
}

func uploadFileToS3(s3Client *s3.Client, bucket, key, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}
	return nil
}

