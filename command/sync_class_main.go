package command

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"fmt"
	"log"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// SyncClassMainCommand xử lý đồng bộ classes_main từ classes
func SyncClassMainCommand() {
	fmt.Println("🔄 Bắt đầu đồng bộ classes_main từ classes...")

	cfg := config.LoadConfig()
	if err := db.ConnectPostgres(cfg); err != nil {
		log.Fatalf("❌ Lỗi kết nối database: %v", err)
	}

	// Lấy distinct school_id từ classes
	var schoolIds []int64
	if err := db.MasterDB.Model(&models.Class{}).
		Distinct("school_id").
		Where("school_id IS NOT NULL").
		Pluck("school_id", &schoolIds).Error; err != nil {
		log.Fatalf("❌ Lỗi lấy school_id: %v", err)
	}

	fmt.Printf("📊 Tìm thấy %d trường học\n", len(schoolIds))

	totalProcessed := 0
	totalCreated := 0
	totalUpdated := 0

	// Xử lý từng school_id
	for _, schoolId := range schoolIds {
		fmt.Printf("\n🏫 Xử lý school_id: %d\n", schoolId)

		// Lấy tất cả name từ classes của school này
		var classNames []string
		if err := db.MasterDB.Model(&models.Class{}).
			Where("school_id = ?", schoolId).
			Pluck("name", &classNames).Error; err != nil {
			log.Printf("⚠️ Lỗi lấy class names cho school_id %d: %v", schoolId, err)
			continue
		}

		// Xử lý tên lớp: bỏ ký tự cuối cùng nếu là chữ cái thường
		// Map từ processed name -> list original names
		processedNamesMap := make(map[string][]string)
		for _, name := range classNames {
			processedName := processClassName(name)
			if processedName != "" {
				processedNamesMap[processedName] = append(processedNamesMap[processedName], name)
			}
		}

		fmt.Printf("  📝 Tìm thấy %d tên lớp sau khi xử lý\n", len(processedNamesMap))

		// Thêm các tên lớp distinct vào classes_main
		for processedName := range processedNamesMap {
			var classMain models.ClassMain
			err := db.MasterDB.Where("name = ? AND school_id = ? AND deleted_at IS NULL", processedName, schoolId).
				First(&classMain).Error

			if err == gorm.ErrRecordNotFound {
				// Chưa tồn tại, tạo mới
				classMain = models.ClassMain{
					Name:      processedName,
					SchoolId:  schoolId,
					CreatedAt: db.MasterDB.NowFunc(),
					UpdatedAt: db.MasterDB.NowFunc(),
				}
				if err := db.MasterDB.Create(&classMain).Error; err != nil {
					log.Printf("⚠️ Lỗi tạo class_main cho '%s': %v", processedName, err)
					continue
				}
				totalCreated++
				fmt.Printf("  ✅ Tạo mới: %s (ID: %d)\n", processedName, classMain.ID)
			} else if err != nil {
				log.Printf("⚠️ Lỗi kiểm tra class_main cho '%s': %v", processedName, err)
				continue
			} else {
				// Đã tồn tại, giữ nguyên (không tạo mới)
				fmt.Printf("  ℹ️ Đã tồn tại, giữ nguyên: %s (ID: %d)\n", processedName, classMain.ID)
			}

			// Gắn class_main_id cho các classes có name sau khi xử lý giống với class_main.name
			// Lấy tất cả classes của school này và cập nhật nếu processed name khớp
			updatedCount := 0
			var allClasses []models.Class
			if err := db.MasterDB.Model(&models.Class{}).
				Where("school_id = ?", schoolId).
				Find(&allClasses).Error; err != nil {
				log.Printf("⚠️ Lỗi lấy classes cho school_id %d: %v", schoolId, err)
				continue
			}

			for _, class := range allClasses {
				// Kiểm tra xem processed name có khớp không
				if processClassName(class.Name) == processedName {
					// Chỉ cập nhật nếu chưa có class_main_id hoặc class_main_id khác
					if class.ClassMainId == 0 || class.ClassMainId != classMain.ID {
						result := db.MasterDB.Model(&models.Class{}).
							Where("id = ?", class.ID).
							Update("class_main_id", classMain.ID)
						if result.Error != nil {
							log.Printf("⚠️ Lỗi cập nhật class_main_id cho class ID %d: %v", class.ID, result.Error)
							continue
						}
						if result.RowsAffected > 0 {
							updatedCount++
						}
					}
				}
			}

			if updatedCount > 0 {
				totalUpdated += updatedCount
				fmt.Printf("  🔗 Cập nhật %d classes với class_main_id = %d\n", updatedCount, classMain.ID)
			}
		}

		totalProcessed++
	}

	fmt.Printf("\n✅ Hoàn thành!\n")
	fmt.Printf("📊 Tổng kết:\n")
	fmt.Printf("  - Số trường đã xử lý: %d\n", totalProcessed)
	fmt.Printf("  - Số class_main đã tạo: %d\n", totalCreated)
	fmt.Printf("  - Số classes đã cập nhật: %d\n", totalUpdated)
}

// processClassName xử lý tên lớp: bỏ ký tự cuối cùng nếu là chữ cái thường
func processClassName(name string) string {
	if name == "" {
		return ""
	}

	// Chuyển thành rune để xử lý Unicode đúng cách
	runes := []rune(name)
	if len(runes) == 0 {
		return ""
	}

	lastChar := runes[len(runes)-1]

	// Nếu ký tự cuối cùng là chữ cái thường, bỏ đi
	if unicode.IsLower(lastChar) && unicode.IsLetter(lastChar) {
		return string(runes[:len(runes)-1])
	}

	return strings.TrimSpace(name)
}
