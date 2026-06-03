package services

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"be-lms/config"
	"be-lms/repositories"

	"github.com/tus/tusd/v2/pkg/filelocker"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

var (
	tusInitOnce    sync.Once
	tusHTTPHandler http.Handler
)

// GetTusHandler initializes (once) and returns the tusd HTTP handler for resumable uploads.
func GetTusHandler() http.Handler {
	tusInitOnce.Do(func() {
		// Determine storage directory under the public disk
		rootPublic := config.Disks[config.Public].Root
		uploadDir := filepath.Join(rootPublic, "uploads", "tus")
		_ = os.MkdirAll(uploadDir, 0755)

		// Log cấu hình upload
		tusConfig := config.GetTUSConfig()
		config.Log.Infof("TUS upload directory: %s", uploadDir)
		config.Log.Infof("TUS chunk size: %d bytes", tusConfig.ChunkSize)
		config.Log.Infof("TUS max upload size: %d bytes", tusConfig.MaxUploadSize)
		config.Log.Infof("TUS upload timeout: %v", tusConfig.UploadTimeout)

		store := filestore.New(uploadDir)
		locker := filelocker.New(uploadDir)

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

		// On complete: rename stored file to original filename and clean .info
		go func(dir string) {
			mediaRepo := repositories.NewMediaRepository()
			config.Log.Info("tus upload cleanup started")
			err := mediaRepo.CleanupOrphanedMedia()

			if err != nil {
				config.Log.Errorf("failed to cleanup orphaned media: %v", err)
			}

			for event := range handler.CompleteUploads {
				config.Log.Infof("tus upload completed: id=%s size=%d", event.Upload.ID, event.Upload.Size)

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

				src := filepath.Join(dir, event.Upload.ID)
				dst := filepath.Join(dir, safeName)

				// Avoid overwrite by appending -<id> if needed
				if _, err := os.Stat(dst); err == nil {
					base := strings.TrimSuffix(safeName, ext)
					dst = filepath.Join(dir, base+"-"+event.Upload.ID+ext)
				}

				if err := os.Rename(src, dst); err != nil {
					config.Log.Errorf("tus rename failed id=%s: %v", event.Upload.ID, err)
				} else {
					infoPath := filepath.Join(dir, event.Upload.ID+".info")
					_ = os.Remove(infoPath)
					config.Log.Infof("tus saved file: %s", dst)

					// Background job: if this is a power_point path, extract and register media
					if strings.EqualFold(event.Upload.MetaData["path"], "power_point") {
						// folder name based on original name without extension + timestamp suffix
						uploadService := NewPowerPointUploadService()
						folderName := getBaseName(dst)
						// Extract
						if res, err := uploadService.UploadAndExtractFromPath(dst, folderName); err != nil {
							config.Log.Errorf("power_point extract failed: %v", err)
						} else {
							// Scan and save medias (multiple packages supported)
							if pkgs, err := uploadService.FindPackages(folderName); err == nil {
								if len(pkgs) == 0 {
									// Fallback to original single behavior (use extract result)
									_, _ = uploadService.SaveMedia(*res)
								} else if len(pkgs) == 1 {
									_, _ = uploadService.SaveMedia(pkgs[0])
								} else {
									for _, p := range pkgs {
										_, _ = uploadService.SaveMedia(p)
									}
								}
							}
						}
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
			}
		}(uploadDir)

		tusHTTPHandler = handler
	})

	return tusHTTPHandler
}

func getBaseName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base)) // Bỏ ".rar"
}
