package services

import (
	"be-lms/prot"
)

type ContestCourseService interface {
	GetContestCourseDetail(courseID int64, isAssigned *bool) (*prot.ContestCourseDetailData, error)
	AssignContestToCourse(contestID int64, courseID int64) (*prot.ContestCourseAssignmentResponse, error)
	RemoveContestFromCourse(contestID int64, courseID int64) (*prot.ContestCourseAssignmentResponse, error)
}

type contestCourseService struct {
	// TODO: Add repositories when implemented
}

func NewContestCourseService() ContestCourseService {
	return &contestCourseService{}
}

func (s *contestCourseService) GetContestCourseDetail(courseID int64, isAssigned *bool) (*prot.ContestCourseDetailData, error) {
	// TODO: Implement logic to get contest course detail
	return &prot.ContestCourseDetailData{}, nil
}

func (s *contestCourseService) AssignContestToCourse(contestID int64, courseID int64) (*prot.ContestCourseAssignmentResponse, error) {
	// TODO: Implement logic to assign contest to course
	return &prot.ContestCourseAssignmentResponse{Success: true, Message: "Contest assigned to course successfully"}, nil
}

func (s *contestCourseService) RemoveContestFromCourse(contestID int64, courseID int64) (*prot.ContestCourseAssignmentResponse, error) {
	// TODO: Implement logic to remove contest from course
	return &prot.ContestCourseAssignmentResponse{Success: true, Message: "Contest removed from course successfully"}, nil
}
