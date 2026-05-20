package services

import (
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/repositories/base"
	"be-cleverschool/requests"
	"be-cleverschool/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type FeedbackService interface {
	GetAll(c *gin.Context) ([]dto.FeedbackResponse, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Feedback, error)
	Create(c *gin.Context, req *prot.FeedbackRequest) (*dto.FeedbackResponse, error)
	Update(c *gin.Context, req *prot.FeedbackRequest) (*dto.FeedbackResponse, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*dto.FeedbackResponse, error)
	UpdateStatus(c *gin.Context, id int64, status *int64, response string, note string, feedbackType *int64) (*prot.Feedback, error)
}

type feedbackService struct {
	repo repositories.FeedbackRepository
}

func NewFeedbackService(r repositories.FeedbackRepository) FeedbackService {
	return &feedbackService{repo: r}
}

func (s *feedbackService) GetAll(c *gin.Context) ([]dto.FeedbackResponse, int64, error) {

	var req requests.GetFeedbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}
	// Lấy user_id và role_id từ token
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		return nil, 0, err
	}

	roleID := utils.GetCurrentRoleId(c)

	var roleIDs []int
	if roleIDsVal, exists := c.Get("roleIDs"); exists {
		if ids, ok := roleIDsVal.([]int); ok {
			roleIDs = ids
		}
	}

	for _, id := range roleIDs {
		if id == models.AdminRoleId {
			roleID = models.AdminRoleId
			break
		}
	}

	// Nếu role_id khác 1 (không phải admin), chỉ lấy feedback của user hiện tại
	if roleID != models.AdminRoleId  {
		req.UserID = &userID
	}

	feedbacks, total, err := s.repo.GetAllWithPaging(&req)
	if err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

func (s *feedbackService) GetByID(c *gin.Context, id int) (*prot.Feedback, error) {
	// Lấy user_id và role_id từ token
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		return nil, err
	}

	var roleIDs []int
	if roleIDsVal, exists := c.Get("roleIDs"); exists {
		if ids, ok := roleIDsVal.([]int); ok {
			roleIDs = ids
		}
	}

	feedback, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, err
	}

	var isAdmin bool
	for _, id := range roleIDs {
		if id == models.AdminRoleId {
			isAdmin = true
			break
		}
	}

	// Nếu role_id khác 1 (không phải admin), kiểm tra xem feedback có phải của user hiện tại không
	if !isAdmin {
		if feedback.UserID == nil || *feedback.UserID != userID {
			return nil, fmt.Errorf("access denied: you can only view your own feedback")
		}
	}

	return dtoToProtoFeedback(feedback), nil
}

func (s *feedbackService) Create(c *gin.Context, req *prot.FeedbackRequest) (*dto.FeedbackResponse, error) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil, nil
	}

	roleID := utils.GetCurrentRoleId(c)
	roleIDInt64 := int64(roleID)

	// Set default status if not provided
	status := req.Status
	if status == 0 {
		defaultStatus := models.FeedbackStatusPending
		status = defaultStatus
	}

	fileUrl := utils.StripDomain(req.FileUrl, models.Storage)
	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	feedback := &models.Feedback{
		UserID:    &userID,
		Content:   req.Content,
		Title:     req.Title,
		FileInfo:  fileInfo,
		RoleID:    &roleIDInt64,
		Status:    &status,
		Response:  req.Response,
		Note:      req.Note,
		Type:      &req.Type,
		CreatedBy: userID,
	}

	err = s.repo.Create(feedback)
	if err != nil {
		return nil, err
	}
	return modelToDtoFeedback(feedback), nil
}

func (s *feedbackService) Update(c *gin.Context, req *prot.FeedbackRequest) (*dto.FeedbackResponse, error) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil, nil
	}

	var roleIDs []int
	if roleIDsVal, exists := c.Get("roleIDs"); exists {
		if ids, ok := roleIDsVal.([]int); ok {
			roleIDs = ids
		}
	}

	// Lấy feedback hiện tại để so sánh và chỉ cập nhật những trường được truyền vào
	existingFeedback, err := s.repo.GetByID(req.Id)
	if err != nil {
		return nil, err
	}

	var isAdmin bool
	for _, id := range roleIDs {
		if id == models.AdminRoleId {
			isAdmin = true
			break
		}
	}

	// Nếu role_id khác 1 (không phải admin), kiểm tra xem feedback có phải của user hiện tại không
	if !isAdmin {
		if existingFeedback.UserID == nil || *existingFeedback.UserID != userID {
			return nil, fmt.Errorf("access denied: you can only update your own feedback")
		}
	}

	// Tạo map updates để chỉ cập nhật những trường được truyền vào
	updates := map[string]interface{}{
		"updated_by": userID,
		"updated_at": time.Now().UTC(),
	}

	// Chỉ cập nhật Content nếu được truyền vào và khác rỗng
	if req.Content != "" {
		updates["content"] = req.Content
	}

	// Chỉ cập nhật Title nếu được truyền vào và khác rỗng
	if req.Title != "" {
		updates["title"] = req.Title
	}

	// Chỉ cập nhật FileInfo nếu FileUrl được truyền vào và khác rỗng
	if req.FileUrl != "" {
		fileUrl := utils.StripDomain(req.FileUrl, models.Storage)
		mediaRepo := repositories.NewMediaRepository()
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
		updates["file_info"] = fileInfo
	}

	// Chỉ cập nhật Status nếu được truyền vào và khác 0
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	// Chỉ cập nhật Response nếu được truyền vào và khác rỗng
	if req.Response != "" {
		updates["response"] = req.Response
	}

	// Chỉ cập nhật Note nếu được truyền vào và khác rỗng
	if req.Note != "" {
		updates["note"] = req.Note
	}

	// Chỉ cập nhật Type nếu được truyền vào và khác 0
	if req.Type != 0 {
		updates["type"] = req.Type
	}

	// Cập nhật feedback với những trường được truyền vào
	err = s.repo.UpdatePartial(req.Id, updates)
	if err != nil {
		return nil, err
	}

	// Lấy feedback đã cập nhật để trả về
	updatedFeedback, err := s.repo.GetByID(req.Id)
	if err != nil {
		return nil, err
	}

	return modelToDtoFeedback(&models.Feedback{
		ID:        updatedFeedback.ID,
		UserID:    updatedFeedback.UserID,
		Content:   updatedFeedback.Content,
		Title:     updatedFeedback.Title,
		FileInfo:  updatedFeedback.FileInfo,
		RoleID:    updatedFeedback.RoleID,
		Status:    updatedFeedback.Status,
		Response:  updatedFeedback.Response,
		Note:      updatedFeedback.Note,
		Type:      updatedFeedback.Type,
		CreatedAt: updatedFeedback.CreatedAt,
		CreatedBy: updatedFeedback.CreatedBy,
		UpdatedAt: updatedFeedback.UpdatedAt,
		UpdatedBy: updatedFeedback.UpdatedBy,
	}), nil
}

func (s *feedbackService) UpdateStatus(c *gin.Context, id int64, status *int64, response string, note string, feedbackType *int64) (*prot.Feedback, error) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil, nil
	}

	var roleIDs []int
	if roleIDsVal, exists := c.Get("roleIDs"); exists {
		if ids, ok := roleIDsVal.([]int); ok {
			roleIDs = ids
		}
	}

	var isAdmin bool
	for _, id := range roleIDs {
		if id == models.AdminRoleId {
			isAdmin = true
			break
		}
	}

	// Nếu role_id khác 1 (không phải admin), kiểm tra xem feedback có phải của user hiện tại không
	if !isAdmin {
		existingFeedback, err := s.repo.GetByID(id)
		if err != nil {
			return nil, err
		}
		if existingFeedback.UserID == nil || *existingFeedback.UserID != userID {
			return nil, fmt.Errorf("access denied: you can only update status of your own feedback")
		}
	}

	// Tạo map updates để chỉ cập nhật những trường được truyền vào
	updates := map[string]interface{}{
		"updated_by": userID,
		"updated_at": time.Now().UTC(),
	}

	// Chỉ cập nhật Status nếu được truyền vào
	if status != nil {
		updates["status"] = *status
	}

	// Chỉ cập nhật Response nếu được truyền vào và khác rỗng
	if response != "" {
		updates["response"] = response
	}

	// Chỉ cập nhật Note nếu được truyền vào và khác rỗng
	if note != "" {
		updates["note"] = note
	}

	// Chỉ cập nhật Type nếu được truyền vào
	if feedbackType != nil {
		updates["type"] = *feedbackType
	}

	// Cập nhật feedback với những trường được truyền vào
	err = s.repo.UpdatePartial(id, updates)
	if err != nil {
		return nil, err
	}

	// Lấy feedback đã update để trả về
	updatedFeedback, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return dtoToProtoFeedback(updatedFeedback), nil
}

func (s *feedbackService) Delete(c *gin.Context, id int) error {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		return err
	}

	var roleIDs []int
	if roleIDsVal, exists := c.Get("roleIDs"); exists {
		if ids, ok := roleIDsVal.([]int); ok {
			roleIDs = ids
		}
	}

	var isAdmin bool
	for _, id := range roleIDs {
		if id == models.AdminRoleId {
			isAdmin = true
			break
		}
	}

	// Nếu role_id khác 1 (không phải admin), kiểm tra xem feedback có phải của user hiện tại không
	if !isAdmin {
		existingFeedback, err := s.repo.GetByID(int64(id))
		if err != nil {
			return err
		}
		if existingFeedback.UserID == nil || *existingFeedback.UserID != userID {
			return fmt.Errorf("access denied: you can only delete your own feedback")
		}
	}

	deletedBy := utils.GetCurrentUserId(c)
	return s.repo.Delete(int64(id), int64(deletedBy))
}

func (s *feedbackService) Restore(c *gin.Context, id int) (*dto.FeedbackResponse, error) {
	baseRepo := base.NewBaseRepository[*models.Feedback]()
	baseRepo.SetContext(c)
	feedback, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	formatFeedback := modelToDtoFeedback(*feedback)

	return formatFeedback, nil
}

func modelToDtoFeedback(feedback *models.Feedback) *dto.FeedbackResponse {
	var roleName string

	if len(feedback.User.Roles) > 0 {
		roleName = feedback.User.Roles[0].Name
	}

	return &dto.FeedbackResponse{
		ID:        feedback.ID,
		UserID:    feedback.UserID,
		Content:   feedback.Content,
		Title:     feedback.Title,
		FileInfo:  feedback.FileInfo,
		RoleID:    feedback.RoleID,
		Status:    feedback.Status,
		Response:  feedback.Response,
		Note:      feedback.Note,
		Type:      feedback.Type,
		CreatedAt: feedback.CreatedAt,
		CreatedBy: feedback.CreatedBy,
		UpdatedAt: feedback.UpdatedAt,
		UpdatedBy: feedback.UpdatedBy,
		Username:  feedback.User.Username,
		Name:      feedback.User.Name,
		RoleName:  roleName,
	}
}

func dtoToProtoFeedback(feedback *dto.FeedbackResponse) *prot.Feedback {
	return &prot.Feedback{
		Id:        feedback.ID,
		UserId:    utils.Int64OrZero(feedback.UserID),
		Content:   feedback.Content,
		Title:     feedback.Title,
		FileUrl:   utils.StaticURL(feedback.FileInfo.Path, models.Storage),
		RoleId:    utils.Int64OrZero(feedback.RoleID),
		Status:    utils.Int64OrZero(feedback.Status),
		Response:  feedback.Response,
		Note:      feedback.Note,
		Type:      utils.Int64OrZero(feedback.Type),
		CreatedAt: feedback.CreatedAt.Unix(),
		CreatedBy: feedback.CreatedBy,
		UpdatedAt: feedback.UpdatedAt.Unix(),
		UpdatedBy: feedback.UpdatedBy,
		Username:  feedback.Username,
		Name:      feedback.Name,
		RoleName:  feedback.RoleName,
	}
}

