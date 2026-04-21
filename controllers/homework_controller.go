package controllers

import (
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"net/http"
	"strconv"

	"be-lms/database/db"
	"be-lms/models"

	"github.com/gin-gonic/gin"
)

type HomeworkController struct {
	svc services.HomeworkService
	*GenericController[models.Homework, prot.Homework, *prot.HomeworkRequest]
}

func NewHomeworkController(service services.HomeworkService) *HomeworkController {
	homeworkResource := resources.NewHomeworkResource()
	homeworkResourceAdapter := NewHomeworkResourceAdapter(homeworkResource)

	genericController := NewGenericController(
		service,
		homeworkResourceAdapter,
		func() *prot.HomeworkRequest {
			return &prot.HomeworkRequest{}
		},
		func(homeworks []*prot.Homework, totalCount uint64) interface{} {
			return &prot.HomeworkListResponse{
				Homeworks:    homeworks,
				Total: int64(totalCount),
			}
		},
	)

	return &HomeworkController{
		GenericController: genericController,
		svc:           service,
	}
}

func (ctl *HomeworkController) UpdateAllHomeworkTotalQuestions(c *gin.Context) {
	var homeworks []models.Homework
	if err := db.MasterDB.Where("deleted_at IS NULL").Find(&homeworks).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	updated := 0
	for _, hw := range homeworks {
		err := services.UpdateHomeworkTotalQuestions(hw.ID)
		if err == nil {
			updated++
		}
	}
	c.JSON(200, gin.H{"updated": updated, "total": len(homeworks)})
}

func (ctl *HomeworkController) Cloned(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	homework, err := ctl.svc.Cloned(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.data_existed", http.StatusNotFound)
		return
	}

	utils.Respond(c, homework, err, "")
}

func (ctl *HomeworkController) Assigned(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	assigned, err := ctl.svc.Assigned(c, id)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, assigned, err, "")
}

func (ctl *HomeworkController) AssignedLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	lessons, err := ctl.svc.AssignedLesson(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, lessons, err, "")
}

func (ctl *HomeworkController) StudentsDoing(c *gin.Context) {
	result, err := ctl.svc.GetStudentsDoing(c)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, result, nil, "")
}

// GetByID override để xử lý permission denied cho học sinh
func (ctl *HomeworkController) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := ctl.svc.GetByID(c, id)
	if err != nil {
		// Kiểm tra nếu là lỗi permission denied
		if err.Error() == "don't have permission to access this homework" {
			utils.Respond(c, nil, err, "don't have permission to access this homework", http.StatusUnauthorized)
			return
		}
		// Các lỗi khác
		utils.Respond(c, nil, err, "messages.error_get_data", http.StatusNotFound)
		return
	}

	utils.Respond(c, item, err, "")
}