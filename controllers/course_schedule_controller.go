package controllers

import (
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CourseScheduleController struct {
	svc services.CourseScheduleService
}

func NewCourseScheduleController() *CourseScheduleController {
	return &CourseScheduleController{
		svc: services.NewCourseScheduleService(),
	}
}

func (ctl *CourseScheduleController) AssignParent(c *gin.Context) {
	var req requests.CourseScheduleAssignParentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse course_ids từ string (cách nhau bởi dấu phẩy) thành array
	// Cho phép course_ids rỗng - khi đó chỉ cập nhật course cha thành parent_course_id = 0
	var courseIDs []int64
	if req.CourseIDs != "" {
		courseIDStrs := strings.Split(req.CourseIDs, ",")
		for _, idStr := range courseIDStrs {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				utils.Respond(c, nil, err, "course_ids không hợp lệ", http.StatusBadRequest)
				return
			}
			if id <= 0 {
				utils.Respond(c, nil, nil, "course_ids phải lớn hơn 0", http.StatusBadRequest)
				return
			}
			courseIDs = append(courseIDs, id)
		}
	}

	resp, err := ctl.svc.AssignParentAndCopySchedule(courseIDs, req.ParentCourseID)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, resp, nil, "")
}

func (ctl *CourseScheduleController) SyncFamily(c *gin.Context) {
	var req requests.CourseScheduleSyncFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate course_id
	if req.CourseID <= 0 {
		utils.Respond(c, nil, nil, "course_id phải lớn hơn 0", http.StatusBadRequest)
		return
	}

	// Thiết lập SSE headers
	w := c.Writer

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	type progressEvent struct {
		TargetCourseID   int64  `json:"target_course_id"`
		Status           string `json:"status"` // "success" | "error"
		Message          string `json:"message"`
		TotalCourses     int    `json:"total_courses"`     // Tổng số khóa cần copy
		CompletedCourses int    `json:"completed_courses"` // Số khóa đã copy xong
	}

	type doneEvent struct {
		Status          string  `json:"status"` // "done"
		SourceCourseID  int64   `json:"source_course_id"`
		SyncedCourseIDs []int64 `json:"synced_course_ids"`
		SyncedCount     int32   `json:"synced_count"`
	}

	sendEvent := func(eventName string, data interface{}) {
		payload, _ := json.Marshal(data)
		_, _ = w.Write([]byte("event: " + eventName + "\n"))
		_, _ = w.Write([]byte("data: " + string(payload) + "\n\n"))
		flusher.Flush()
	}

	// Lấy thông tin family: root + children
	familyRepo := repositories.NewCourseFamilyRepository()
	basic, err := familyRepo.GetCourseBasic(req.CourseID)
	if err != nil || basic == nil {
		sendEvent("error", progressEvent{
			TargetCourseID:   0,
			Status:           "error",
			Message:          "Không tìm thấy khóa học nguồn",
			TotalCourses:     0,
			CompletedCourses: 0,
		})
		return
	}

	rootCourseID := basic.ID
	if basic.ParentCourseID != 0 {
		rootCourseID = basic.ParentCourseID
	}

	children, err := familyRepo.GetChildren(rootCourseID)
	if err != nil {
		sendEvent("error", progressEvent{
			TargetCourseID:   0,
			Status:           "error",
			Message:          "Không thể lấy danh sách khóa trong family",
			TotalCourses:     0,
			CompletedCourses: 0,
		})
		return
	}

	var targetCourseIDs []int64
	targetCourseIDs = append(targetCourseIDs, rootCourseID)
	for _, child := range children {
		if child.ID > 0 {
			targetCourseIDs = append(targetCourseIDs, child.ID)
		}
	}

	// Lọc ra các khóa cần copy (loại bỏ khóa nguồn)
	var coursesToSync []int64
	for _, targetID := range targetCourseIDs {
		if targetID != req.CourseID {
			coursesToSync = append(coursesToSync, targetID)
		}
	}

	totalCourses := len(coursesToSync)

	if totalCourses == 0 {
		// Chỉ có khóa nguồn, không có gì để sync
		sendEvent("done", doneEvent{
			Status:          "done",
			SourceCourseID:  req.CourseID,
			SyncedCourseIDs: []int64{},
			SyncedCount:     0,
		})
		return
	}

	// Dùng shared copy service để copy schedules
	copyService := services.NewLessonScheduleCopySharedService()
	var syncedCourseIDs []int64
	completedCount := 0

	// Context để kiểm tra client disconnect
	ctx := c.Request.Context()

	// Ticker để gửi keep-alive mỗi 30 giây (tránh timeout)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Channel để báo khi copy xong
	doneChan := make(chan bool)

	// Goroutine để gửi keep-alive
	go func() {
		for {
			select {
			case <-ticker.C:
				// Gửi keep-alive comment (SSE format)
				_, _ = w.Write([]byte(": keep-alive\n\n"))
				flusher.Flush()
			case <-doneChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		sem = make(chan struct{}, 5) // max 5 concurrent jobs
	)

	for _, targetID := range coursesToSync {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)

		go func(targetID int64) {
			defer wg.Done()

			// limit concurrency
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := copyService.CopyLessonSchedulesBetweenCourses(req.CourseID, targetID)

			mu.Lock()
			completedCount++
			defer mu.Unlock()

			if err != nil {
				sendEvent("progress", progressEvent{
					TargetCourseID:   targetID,
					Status:           "error",
					Message:          err.Error(),
					TotalCourses:     totalCourses,
					CompletedCourses: completedCount,
				})
				return
			}

			if !result.Success {
				sendEvent("progress", progressEvent{
					TargetCourseID:   targetID,
					Status:           "error",
					Message:          result.Message,
					TotalCourses:     totalCourses,
					CompletedCourses: completedCount,
				})
				return
			}

			syncedCourseIDs = append(syncedCourseIDs, targetID)

			sendEvent("progress", progressEvent{
				TargetCourseID:   targetID,
				Status:           "success",
				Message:          result.Message,
				TotalCourses:     totalCourses,
				CompletedCourses: completedCount,
			})
		}(targetID)
	}

	wg.Wait()

	// Dừng keep-alive ticker
	close(doneChan)

	// Gửi event done
	sendEvent("done", doneEvent{
		Status:          "done",
		SourceCourseID:  req.CourseID,
		SyncedCourseIDs: syncedCourseIDs,
		SyncedCount:     int32(len(syncedCourseIDs)),
	})

	// Đóng kết nối SSE sau khi gửi event done
	// Gửi thêm một dòng trống để đảm bảo client nhận được event done trước khi đóng
	_, _ = w.Write([]byte("\n"))
	flusher.Flush()

	// Return để đóng kết nối (Gin sẽ tự động đóng response writer)
	return
}
