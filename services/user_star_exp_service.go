package services

import (
	"be-Clever School/models"
	"be-Clever School/repositories"
	"fmt"
	"time"
)

type UserStarExpService interface {
	UpdateStarAndExp(userID int64, changeStar int64, changeExp float64, note string) error
}

type userStarExpService struct {
	repo repositories.UserStarExpRepository
}

func NewUserStarExpService() UserStarExpService {
	return &userStarExpService{
		repo: repositories.NewUserStarExpRepository(),
	}
}

// UpdateStarAndExp cập nhật tổng exp và tổng star vào user_star_exp
// changeStar và changeExp có thể dương (tăng) hoặc âm (giảm)
// note là mô tả cho thay đổi này (ví dụ: "homework 123", "chấm điểm lại", ...)
func (s *userStarExpService) UpdateStarAndExp(userID int64, changeStar int64, changeExp float64, note string) error {
	// Lấy record hiện tại (is_current = true) của user
	currentRecord, err := s.repo.GetCurrentByUserID(userID)
	if err != nil {
		return fmt.Errorf("không thể lấy record hiện tại: %w", err)
	}

	var totalStar int64
	var totalExp float64

	if currentRecord == nil {
		// Chưa có record nào, bắt đầu từ 0
		totalStar = changeStar
		totalExp = changeExp
	} else {
		// Đã có record, tính toán dựa trên record hiện tại
		totalStar = currentRecord.TotalStar + changeStar
		totalExp = currentRecord.TotalExp + changeExp
	}

	// Đảm bảo total không âm
	if totalStar < 0 {
		totalStar = 0
	}
	if totalExp < 0 {
		totalExp = 0
	}

	// Đặt tất cả record của user thành is_current = false
	if err := s.repo.UpdateIsCurrentForUser(userID, false); err != nil {
		return fmt.Errorf("không thể cập nhật is_current: %w", err)
	}

	// Tạo record mới với is_current = true
	newRecord := &models.UserStarExp{
		UserID:      userID,
		TotalStar:   totalStar,
		ChangeStar:  changeStar,
		TotalExp:    totalExp,
		ChangeExp:   changeExp,
		IsCurrent:   true,
		CreatedAt:   time.Now(),
		Description: note,
	}

	if err := s.repo.Create(newRecord); err != nil {
		return fmt.Errorf("không thể tạo record mới: %w", err)
	}

	return nil
}

