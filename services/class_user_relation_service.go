package services

import (
	"be-lms/prot"
	"be-lms/repositories"

	"github.com/gin-gonic/gin"
)

type ClassUserRelationService interface {
	UpdateUserClassRelation(c *gin.Context, classID int64, studentIDs []int64, updatedBy int64) (*prot.ClassUserRelationResponse, error)
}

type classUserRelationService struct {
	repo repositories.ClassUserRelationRepository
}

func NewClassUserRelationService() ClassUserRelationService {
	return &classUserRelationService{
		repo: repositories.NewClassUserRelationRepository(),
	}
}

func (s *classUserRelationService) UpdateUserClassRelation(c *gin.Context, classID int64, studentIDs []int64, updatedBy int64) (*prot.ClassUserRelationResponse, error) {
	err := s.repo.UpdateUserClassRelation(classID, studentIDs, updatedBy)
	if err != nil {
		return &prot.ClassUserRelationResponse{
			Success: false,
			Message: "Cập nhật thất bại: " + err.Error(),
		}, nil
	}

	return &prot.ClassUserRelationResponse{
		Success: true,
		Message: "Cập nhật thành công",
	}, nil
}
