package services

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"be-lms/config"
	"be-lms/repositories"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tus/tusd/v2/pkg/filelocker"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"
)

var (
	tusInitOnce    sync.Once
	tusHTTPHandler http.Handler
)

// GetTusHandler initializes (once) and returns the tusd HTTP handler for resumable uploads.
func GetTusHandler() http.Handler {
	tusInitOnce.Do(func() {
		tusConfig := config.GetTUSConfig()
		s3Disk := config.Disks[config.S3]

		// Log cấu hình upload
		config.Log.Infof("TUS storage: S3")
		config.Log.Infof("TUS S3 bucket: %s", s3Disk.Bucket)
		config.Log.Infof("TUS chunk size: %d bytes", tusConfig.ChunkSize)
		config.Log.Infof("TUS max upload size: %d bytes", tusConfig.MaxUploadSize)
		config.Log.Infof("TUS upload timeout: %v", tusConfig.UploadTimeout)

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
			config.Log.Errorf("failed to load AWS config: %v", err)
			tusHTTPHandler = http.NotFoundHandler()
			return
		}

		client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = s3Disk.PathStyle
		})

		// Create S3 store
		store := s3store.New(s3Disk.Bucket, client)
		store.ObjectPrefix = "uploads/tus"

		rootPublic := config.Disks[config.Public].Root
		lockDir := filepath.Join(rootPublic, "uploads", "tus-locks")
		_ = os.MkdirAll(lockDir, 0755)
		locker := filelocker.New(lockDir)

		composer := tusd.NewStoreComposer()
		store.UseIn(composer)
		locker.UseIn(composer)

		handler, err := tusd.NewHandler(tusd.Config{
			BasePath:                "/api/tus-uploads/",
			StoreComposer:           composer,
			NotifyCompleteUploads:   true,
			RespectForwardedHeaders: true,
		})
		if err != nil {
			config.Log.Errorf("failed to init tus handler: %v", err)
			tusHTTPHandler = http.NotFoundHandler()
			return
		}

		// On complete: process uploaded file from S3
		go func() {
			mediaRepo := repositories.NewMediaRepository()
			config.Log.Info("tus upload cleanup started")
			err := mediaRepo.CleanupOrphanedMedia()

			if err != nil {
				config.Log.Errorf("failed to cleanup orphaned media: %v", err)
			}

			for event := range handler.CompleteUploads {
				originalName := event.Upload.MetaData["filename"]
				safeName := filepath.Base(originalName)
				if safeName == "." || safeName == "" {
					safeName = event.Upload.ID
				}
				// If no extension was provided, default to .rar as requested
				ext := filepath.Ext(safeName)
				if ext == "" {
					ext = ".rar"
					safeName = safeName + ext
				}

				baseUploadID := event.Upload.ID
				if idx := strings.Index(baseUploadID, "+"); idx > 0 {
					baseUploadID = baseUploadID[:idx]
				}
				s3Key := filepath.ToSlash(filepath.Join(store.ObjectPrefix, baseUploadID))

				// Background job: if this is a power_point path, download from S3, extract and register media
				if strings.EqualFold(event.Upload.MetaData["path"], "power_point") {
					go func(uploadID, baseID, key string) {
						uploadService := NewPowerPointUploadService()
						folderName := getBaseName(safeName)

						// Download from S3 to a temp file
						tempDir := filepath.Join("temp", "tus-downloads")
						_ = os.MkdirAll(tempDir, 0755)
						tempFile := filepath.Join(tempDir, uploadID+ext)
						defer func() {
							if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
								config.Log.Infof("failed to remove temp file %s: %v", tempFile, err)
							}
						}()

						time.Sleep(3 * time.Second)

						// Download object from S3 with retry (S3 eventual consistency)
						var downloadResult *s3.GetObjectOutput
						var err error
						maxRetries := 10
						retryDelay := 1 * time.Second

						// First, try to check if file exists using HeadObject
						for i := 0; i < maxRetries; i++ {
							if i > 0 {
								time.Sleep(retryDelay)
								retryDelay = time.Duration(float64(retryDelay) * 1.5) // Exponential backoff
							}

							// Try HeadObject first to verify file exists
							_, headErr := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
								Bucket: aws.String(s3Disk.Bucket),
								Key:    aws.String(key),
							})

							if headErr != nil {
								config.Log.Infof("Attempt %d/%d: file not found in S3 (key: %s): %v", i+1, maxRetries, key, headErr)
								err = headErr // Set err so we know it failed
								continue
							}

							// File exists, now try to download
							downloadResult, err = client.GetObject(context.TODO(), &s3.GetObjectInput{
								Bucket: aws.String(s3Disk.Bucket),
								Key:    aws.String(key),
							})

							if err == nil {
								break
							}

							config.Log.Infof("Attempt %d/%d: failed to download from S3 (key: %s): %v", i+1, maxRetries, key, err)
						}

						// If still failed after retries, try to list objects to find the actual key
						if downloadResult == nil {
							config.Log.Infof("Failed to download after %d retries, listing objects in prefix to debug...", maxRetries)

							// Try multiple prefix strategies to find the file
							// Use baseID (part before '+') since s3store only stores with the base ID
							prefixesToTry := []string{
								filepath.ToSlash(store.ObjectPrefix),                        // uploads/tus
								filepath.ToSlash(filepath.Join(store.ObjectPrefix, baseID)), // uploads/tus/{baseID}
								"uploads/", // uploads/
								filepath.ToSlash(filepath.Join("uploads", "tus")), // uploads/tus (without store.ObjectPrefix)
							}

							found := false
							for _, prefixToSearch := range prefixesToTry {
								if found {
									break
								}
								listOutput, listErr := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
									Bucket: aws.String(s3Disk.Bucket),
									Prefix: aws.String(prefixToSearch),
								})
								if listErr != nil {
									config.Log.Warnf("Failed to list objects with prefix %s: %v", prefixToSearch, listErr)
									continue
								}

								if listOutput != nil {
									// Log first few objects for debugging
									maxLog := 10
									for i, obj := range listOutput.Contents {
										if i < maxLog {
											config.Log.Infof("  Object[%d]: %s (size: %d)", i, *obj.Key, *obj.Size)
										}
									}
									if len(listOutput.Contents) > maxLog {
										config.Log.Infof("  ... and %d more objects", len(listOutput.Contents)-maxLog)
									}

									// Find objects that contain the base upload ID (s3store uses base ID, not full ID)
									for _, obj := range listOutput.Contents {
										if strings.Contains(*obj.Key, baseID) && !strings.Contains(*obj.Key, ".info") {
											downloadResult, err = client.GetObject(context.TODO(), &s3.GetObjectInput{
												Bucket: aws.String(s3Disk.Bucket),
												Key:    obj.Key,
											})
											if err == nil {
												found = true
												break
											} else {
												config.Log.Infof("Failed to download with alternative key %s: %v", *obj.Key, err)
											}
										}
									}
								}
							}

							if !found {
								config.Log.Infof("Could not find file with baseID %s in any of the searched prefixes", baseID)
							}
						}

						if downloadResult == nil {
							config.Log.Errorf("failed to download from S3 after %d retries and listing (key: %s): %v", maxRetries, key, err)
							return
						}

						defer downloadResult.Body.Close()

						// Save to temp file and ensure it's flushed
						outFile, err := os.Create(tempFile)
						if err != nil {
							config.Log.Errorf("failed to create temp file: %v", err)
							return
						}

						if _, err := io.Copy(outFile, downloadResult.Body); err != nil {
							outFile.Close()
							config.Log.Errorf("failed to write temp file: %v", err)
							return
						}

						// Ensure file is flushed and closed before extract
						if err := outFile.Sync(); err != nil {
							config.Log.Infof("failed to sync temp file: %v", err)
						}
						if err := outFile.Close(); err != nil {
							config.Log.Errorf("failed to close temp file: %v", err)
							return
						}

						// Extract
						if res, err := uploadService.UploadAndExtractFromPath(tempFile, folderName); err != nil {
							config.Log.Errorf("power_point extract failed: %v", err)
						} else {
							// Scan and save medias (multiple packages supported)
							if pkgs, err := uploadService.FindPackages(folderName); err == nil {
								if len(pkgs) == 0 {
									// Fallback to original single behavior (use extract result)
									if _, err := uploadService.SaveMedia(*res); err != nil {
										config.Log.Errorf("failed to save media: %v", err)
									}
								} else if len(pkgs) == 1 {
									if _, err := uploadService.SaveMedia(pkgs[0]); err != nil {
										config.Log.Errorf("failed to save media: %v", err)
									}
								} else {
									for _, p := range pkgs {
										if _, err := uploadService.SaveMedia(p); err != nil {
											config.Log.Errorf("failed to save media for package %s: %v", p.FolderName, err)
										}
									}
								}
							} else {
								config.Log.Errorf("failed to find packages: %v", err)
							}
						}
					}(event.Upload.ID, baseUploadID, s3Key)
				}

				// Clear duplicate media
				config.Log.Info("Clearing duplicate media...")
				err := mediaRepo.ClearDuplicateMedia()

				if err != nil {
					config.Log.Errorf("failed to clear duplicate media: %v", err)
				} else {
					config.Log.Info("Duplicate media cleared.")
				}
			}
		}()

		tusHTTPHandler = handler
	})

	return tusHTTPHandler
}

func getBaseName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base)) // Bỏ ".rar"
}
