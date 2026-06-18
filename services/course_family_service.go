package services

import (
	"be-lms/models"
	"be-lms/prot"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func courseToFamilyItem(course *models.Course) *prot.CourseFamilyItem {
	if course == nil {
		return nil
	}
	schoolName := ""
	if len(course.Schools) > 0 {
		schoolName = course.Schools[0].Name
	}
	return &prot.CourseFamilyItem{
		Id:          course.ID,
		Name:        course.Name,
		ObjectTitle: course.ObjectTitle,
		SchoolName:  schoolName,
	}
}

func (s *courseService) GetCourseFamily(c *gin.Context, courseID int64) (*prot.CourseFamilyResponse, error) {
	if _, err := s.repo.FindByID(int(courseID)); err != nil {
		return nil, err
	}

	parentID := s.repo.GetCloneParentID(courseID)
	templateID := courseID
	if parentID > 0 {
		templateID = parentID
	}

	parentCourses, err := s.repo.FindCoursesByIDs([]int64{templateID})
	if err != nil {
		return nil, err
	}

	childIDs := s.repo.GetCloneIds(templateID)
	filteredChildIDs := make([]int64, 0, len(childIDs))
	for _, id := range childIDs {
		if id != templateID {
			filteredChildIDs = append(filteredChildIDs, id)
		}
	}

	childrenCourses, err := s.repo.FindCoursesByIDs(filteredChildIDs)
	if err != nil {
		return nil, err
	}

	resp := &prot.CourseFamilyResponse{
		Children: make([]*prot.CourseFamilyItem, 0, len(childrenCourses)),
	}
	if len(parentCourses) > 0 {
		resp.Parent = courseToFamilyItem(&parentCourses[0])
	}
	for i := range childrenCourses {
		if item := courseToFamilyItem(&childrenCourses[i]); item != nil {
			resp.Children = append(resp.Children, item)
		}
	}

	return resp, nil
}

func (s *courseService) familySyncTargetIDs(courseID int64) ([]int64, error) {
	if _, err := s.repo.FindByID(int(courseID)); err != nil {
		return nil, err
	}

	parentID := s.repo.GetCloneParentID(courseID)
	templateID := courseID
	if parentID > 0 {
		templateID = parentID
	}

	childIDs := s.repo.GetCloneIds(templateID)
	targets := make([]int64, 0, len(childIDs))
	for _, id := range childIDs {
		if id != courseID {
			targets = append(targets, id)
		}
	}
	return targets, nil
}

func (s *courseService) SyncAllFamilyCourses(c *gin.Context, courseID int64) error {
	targets, err := s.familySyncTargetIDs(courseID)
	if err != nil {
		return err
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	total := len(targets)
	if total == 0 {
		writeSSE(c, "done", map[string]interface{}{
			"status":           "done",
			"source_course_id": courseID,
			"synced_course_ids": []int64{},
			"synced_count":     0,
		})
		return nil
	}

	syncedIDs := make([]int64, 0, total)
	for i, targetID := range targets {
		result := s.CopySchedule(targetID, courseID, 0)
		status := "success"
		message := result.CopyScheduleMessage
		if !result.CopyScheduleSuccess {
			status = "error"
			writeSSE(c, "error", map[string]interface{}{
				"message": fmt.Sprintf("Khóa học %d: %s", targetID, message),
			})
			return nil
		}
		syncedIDs = append(syncedIDs, targetID)
		writeSSE(c, "progress", map[string]interface{}{
			"target_course_id":   targetID,
			"status":             status,
			"message":            message,
			"total_courses":      total,
			"completed_courses":  i + 1,
		})
	}

	writeSSE(c, "done", map[string]interface{}{
		"status":            "done",
		"source_course_id":  courseID,
		"synced_course_ids": syncedIDs,
		"synced_count":      len(syncedIDs),
	})
	return nil
}

func writeSSE(c *gin.Context, event string, payload map[string]interface{}) {
	data, _ := json.Marshal(payload)
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, string(data))
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
