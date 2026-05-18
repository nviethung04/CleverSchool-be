package base

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type MediaSync[T any] struct{}

func NewMediaSync[T any]() *MediaSync[T] {
	return &MediaSync[T]{}
}

func (ms *MediaSync[T]) SyncMediaInfo() error {
	var items []T
	err := db.MasterDB.Find(&items).Error
	if err != nil {
		config.Log.Error("Failed to get all items:", err)
		return err
	}

	fields := []struct {
		urlFieldName       string
		mediaInfoFieldName string
		dbColumnName       string
	}{
		{"FileUrl", "FileInfo", "file_info"},
		{"MatchingFileUrl", "MatchingFileInfo", "matching_file_info"},
		{"Avatar", "AvatarInfo", "avatar_info"},
		{"Image", "ImageInfo", "image_info"},
		{"CoverImage", "CoverImageInfo", "cover_image_info"},
		{"Link", "LinkInfo", "link_info"},
		{"Logo", "LogoInfo", "logo_info"},
		{"FileURL", "FileInfo", "file_info"},
	}

	for _, i := range items {
		reflectValue := reflect.ValueOf(i)
		if reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}

		idField := reflectValue.FieldByName("ID")
		var id any
		if idField.IsValid() {
			id = idField.Interface()
		}

		for _, f := range fields {
			urlField := reflectValue.FieldByName(f.urlFieldName)
			mediaInfoField := reflectValue.FieldByName(f.mediaInfoFieldName)

			if urlField.IsValid() && mediaInfoField.IsValid() && (urlField.Kind() == reflect.String || urlField.Kind() == reflect.Ptr) {
				fileUrl := urlField.String()
				fileUrl = StripDomain(fileUrl, models.Storage)

				newMediaInfo := models.SyncMediaInfoByPath(db.MasterDB, fileUrl)
				oldMediaInfo := mediaInfoField.Interface()

				if !reflect.DeepEqual(newMediaInfo, oldMediaInfo) {
					err := db.MasterDB.Session(&gorm.Session{SkipHooks: true}).
						Model(i).Update(f.dbColumnName, newMediaInfo).Error

					if err != nil {
						config.Log.Errorf("Failed to update %s for item ID %v: %v", f.dbColumnName, id, err)
					}
				}
			}
		}
	}

	return nil
}

func (ms *MediaSync[T]) SyncMediaInfoForRecentlyUpdatedMedia(within time.Duration) error {
	fieldsObject := []struct {
		mediaInfoFieldName string
		dbColumnName       string
	}{
		{"FileInfo", "file_info"},
		{"MatchingFileInfo", "matching_file_info"},
		{"AvatarInfo", "avatar_info"},
		{"ImageInfo", "image_info"},
		{"CoverImageInfo", "cover_image_info"},
		{"LinkInfo", "link_info"},
		{"LogoInfo", "logo_info"},
	}

	fieldsArray := []struct {
		mediaInfoFieldName string
		dbColumnName       string
	}{
		{"FileInfos", "file_infos"},
	}

	var mediaIDs []int64
	err := db.MasterDB.
		Model(&models.Media{}).
		Where("updated_at >= ?", time.Now().Add(-within)).
		Pluck("id", &mediaIDs).Error
	if err != nil {
		return err
	}
	if len(mediaIDs) == 0 {
		return nil
	}

	// Get table name using GORM's schema parsing
	var model T
	stmt := &gorm.Statement{DB: db.MasterDB}
	if err := stmt.Parse(&model); err != nil {
		config.Log.Error("Failed to parse model schema: ", err)
		return fmt.Errorf("failed to parse model schema: %w", err)
	}
	tableName := stmt.Schema.Table

	// Xử lý fields object
	for _, field := range fieldsObject {
		mediaIDsStr := make([]string, len(mediaIDs))
		for i, v := range mediaIDs {
			mediaIDsStr[i] = fmt.Sprint(v)
		}

		var ids []int64
		err := db.MasterDB.
			Table(tableName).
			Where(fmt.Sprintf("%s->>'id' IN ?", field.dbColumnName), mediaIDsStr).
			Pluck("id", &ids).Error
		if err != nil {
			continue
		}

		if len(ids) == 0 {
			continue
		}

		for _, id := range ids {
			err := ms.SyncMediaInfoByID(id, field.mediaInfoFieldName, field.dbColumnName)
			if err != nil {
				config.Log.Errorf("Failed to sync %s for record ID %d: %v", field.dbColumnName, id, err)
			}
		}
	}

	// Xử lý fields array
	for _, field := range fieldsArray {
		var ids []int64
		err := db.MasterDB.
			Table(tableName).
			Where(fmt.Sprintf(`
				EXISTS (
					SELECT 1 FROM jsonb_array_elements(%s) elem
					WHERE (elem->>'id')::bigint IN ?
				)
			`, field.dbColumnName), mediaIDs).
			Pluck("id", &ids).Error
		if err != nil {
			continue
		}
		if len(ids) == 0 {
			continue
		}

		for _, id := range ids {
			err := ms.SyncMediaInfoByID(id, field.mediaInfoFieldName, field.dbColumnName)
			if err != nil {
				config.Log.Errorf("Failed to sync %s for record ID %d: %v", field.dbColumnName, id, err)
			}
		}
	}

	return nil
}

func (ms *MediaSync[T]) SyncMediaInfoByID(id int64, infoFieldName, dbColumnName string) error {
    var item T
    if err := db.MasterDB.First(&item, id).Error; err != nil {
        return err
    }

    rv := reflect.ValueOf(item)
    if rv.Kind() == reflect.Ptr {
        rv = rv.Elem()
    }
    mediaInfoField := rv.FieldByName(infoFieldName)
    if !mediaInfoField.IsValid() {
        return fmt.Errorf("Field %s not found", infoFieldName)
    }

    switch mediaInfoField.Kind() {
    case reflect.Struct, reflect.Ptr:
        // ---- xử lý object (kể cả *MediaInfo) ----
        mf := mediaInfoField
        if mf.Kind() == reflect.Ptr {
            if mf.IsNil() {
                return nil // không có gì để sync
            }
            mf = mf.Elem()
        }

        if mf.Kind() != reflect.Struct {
            return fmt.Errorf("Invalid MediaInfo type (got %s)", mf.Kind())
        }

        mediaInfo, ok := mf.Interface().(models.MediaInfo)
        if !ok {
            return fmt.Errorf("Invalid MediaInfo type (assert failed)")
        }
        if mediaInfo.Id == 0 {
            return nil
        }

        var media models.Media
        if err := db.MasterDB.First(&media, mediaInfo.Id).Error; err != nil {
            return err
        }
        if media.StaticURL == nil || media.DiskName == nil {
            return fmt.Errorf("Media info incomplete for id %d", media.ID)
        }

        newMediaInfo := models.MediaInfo{
            Id:   media.ID,
            Path: *media.StaticURL,
            Disk: *media.DiskName,
        }

        if !reflect.DeepEqual(mediaInfo, newMediaInfo) {
            if err := db.MasterDB.Model(&item).
                Where("id = ?", id).
                Update(dbColumnName, newMediaInfo).Error; err != nil {
                return err
            }
        }
        return nil

    case reflect.Slice:
        // ---- xử lý mảng (đã OK nhưng bổ sung an toàn cho phần tử pointer) ----
        n := mediaInfoField.Len()
        mediaInfos := make(models.MediaInfos, n)
        for i := 0; i < n; i++ {
            elem := mediaInfoField.Index(i)
            if elem.Kind() == reflect.Ptr {
                if elem.IsNil() {
                    continue
                }
                elem = elem.Elem()
            }
            if elem.Kind() != reflect.Struct {
                return fmt.Errorf("Invalid slice element type, expected struct")
            }
            v, ok := elem.Interface().(models.MediaDetail)
            if !ok {
                return fmt.Errorf("Failed to convert slice element to MediaDetail")
            }
            mediaInfos[i] = v
        }

        changed := false
        for i, info := range mediaInfos {
            if info.Id == 0 {
                continue
            }
            var media models.Media
            if err := db.MasterDB.First(&media, info.Id).Error; err != nil {
                return err
            }
            if media.StaticURL == nil || media.DiskName == nil {
                return fmt.Errorf("Media info incomplete for id %d", media.ID)
            }
            newInfo := models.MediaDetail{
                Id:   media.ID,
                Path: *media.StaticURL,
                Disk: *media.DiskName,
				Type: info.Type,
            }
            if !reflect.DeepEqual(info, newInfo) {
                mediaInfos[i] = newInfo
                changed = true
            }
        }

        if changed {
            if err := db.MasterDB.Model(&item).
                Where("id = ?", id).
                Update(dbColumnName, mediaInfos).Error; err != nil {
                return err
            }
        }
        return nil

    default:
        return fmt.Errorf("Unsupported media info field type %s", mediaInfoField.Kind())
    }
}


func StripDomain(fullURL string, storage string) string {
	disk, ok := config.Disks[storage]
	if !ok {
		config.Log.Info("StripDomain disk error")
		return fullURL
	}

	base := disk.URL

	if fullURL == "" || base == "" {
		return fullURL
	}

	if strings.HasPrefix(fullURL, base) {
		trimmed := strings.TrimPrefix(fullURL, base)
		return strings.TrimPrefix(trimmed, "/")
	}

	return fullURL
}
