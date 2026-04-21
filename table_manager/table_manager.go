package table_manager

import (
	"fmt"
	"time"
)

// GetMonthlyTableName trả về tên bảng theo tháng
// Ví dụ: activity_logs_01_2025, activity_logs_12_2024
func GetMonthlyTableName(baseTableName string, date time.Time) string {
	month := int(date.Month())
	year := date.Year()
	return fmt.Sprintf("%s_%02d_%d", baseTableName, month, year)
}

// GetCurrentMonthTableName trả về tên bảng của tháng hiện tại
func GetCurrentMonthTableName(baseTableName string) string {
	return GetMonthlyTableName(baseTableName, time.Now())
}

// GetPreviousMonthTableName trả về tên bảng của tháng trước
func GetPreviousMonthTableName(baseTableName string) string {
	lastMonth := time.Now().AddDate(0, -1, 0)
	return GetMonthlyTableName(baseTableName, lastMonth)
}

// GetNextMonthTableName trả về tên bảng của tháng sau
func GetNextMonthTableName(baseTableName string) string {
	nextMonth := time.Now().AddDate(0, 1, 0)
	return GetMonthlyTableName(baseTableName, nextMonth)
}

// GetTableNamesForDateRange trả về danh sách tên bảng trong khoảng thời gian
func GetTableNamesForDateRange(baseTableName string, startDate, endDate time.Time) []string {
	var tableNames []string
	
	// Tạo iterator từ startDate đến endDate theo tháng
	current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	
	for current.Before(end) || current.Equal(end) {
		tableNames = append(tableNames, GetMonthlyTableName(baseTableName, current))
		current = current.AddDate(0, 1, 0)
	}
	
	return tableNames
}

// GetTableNamesForLastMonths trả về danh sách tên bảng của N tháng gần nhất
func GetTableNamesForLastMonths(baseTableName string, months int) []string {
	var tableNames []string
	now := time.Now()
	
	for i := 0; i < months; i++ {
		date := now.AddDate(0, -i, 0)
		tableNames = append(tableNames, GetMonthlyTableName(baseTableName, date))
	}
	
	return tableNames
}
