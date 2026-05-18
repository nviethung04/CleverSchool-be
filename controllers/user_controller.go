package controllers

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	svc services.UserService
	*GenericController[models.User, prot.User, *prot.User]
}

// API xuất PDF tài khoản & mật khẩu theo trường/lớp
func (cc *UserController) ExportUsersPDF(c *gin.Context) {
	// Lấy tham số từ query string
	schoolIDStr := c.Query("school_id")
	classIDStr := c.Query("class_id")

	// Validate required parameters
	if schoolIDStr == "" || classIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "school_id và class_id là bắt buộc",
			"data":    nil,
		})
		return
	}

	// Convert to integers
	schoolID, err := strconv.Atoi(schoolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "school_id phải là số nguyên",
			"data":    nil,
		})
		return
	}

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "class_id phải là số nguyên",
			"data":    nil,
		})
		return
	}

	// Call service to export PDF
	pdfData, filename, err := cc.svc.ExportUsersPDF(schoolID, classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Lỗi khi xuất PDF: " + err.Error(),
			"data":    nil,
		})
		return
	}

	// Luôn trả PDF binary trực tiếp để download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

func NewUserController(service services.UserService) *UserController {
	userResource := resources.NewUserResource()
	userResourceAdapter := NewUserResourceAdapter(userResource)

	genericController := NewGenericController(
		service,
		userResourceAdapter,
		func() *prot.User {
			return &prot.User{}
		},
		func(users []*prot.User, totalCount uint64) interface{} {
			return &prot.UsersResponse{
				Users:      users,
				TotalCount: int64(totalCount),
			}
		},
	)

	ctl := &UserController{
		GenericController: genericController,
		svc:               service,
	}

	ctl.GenericController.WithUsedError(func() bool {
		return true
	})

	ctl.GenericController.WithRespondListHook(func(c *gin.Context, items []models.User, totalCount int64, err error) {
		ctl.RespondList(c, items, totalCount, err)
	})

	return ctl
}

func (ctl *UserController) GetStudentsByParent(c *gin.Context) {
	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.Atoi(parentIDStr)
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid")
		return
	}

	data, err := ctl.svc.GetAllStudentsByParentID(parentID)
	utils.Respond(c, &prot.ListUser{Users: data}, err, "")
}

func (uc *UserController) GetMyProfile(c *gin.Context) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}
	user, err := uc.svc.GetByID(c, int(userID))
	if err != nil {
		utils.Respond(c, nil, err, "messages.data_existed", http.StatusNotFound)
		return
	}

	// Lấy total_star và total_exp từ user_star_exp
	userStarExpRepo := repositories.NewUserStarExpRepository()
	userStarExp, err := userStarExpRepo.GetCurrentByUserID(userID)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Set total_star và total_exp vào response
	if userStarExp != nil {
		user.TotalStar = userStarExp.TotalStar
		user.TotalExp = userStarExp.TotalExp
	} else {
		// Nếu chưa có record, set mặc định là 0
		user.TotalStar = 0
		user.TotalExp = 0
	}

	utils.Respond(c, user, err, "")
}

func (sc *UserController) UpdateProfile(c *gin.Context) {
	idVal, ok := c.Get("userID")
	if !ok {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_record_update")), "messages.no_record_update")
		return
	}

	var id int64
	switch v := idVal.(type) {
	case int64:
		id = v
	case int:
		id = int64(v)
	case float64:
		id = int64(v)
	default:
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_record_update")), "messages.no_record_update")
		return
	}

	req, err, message := utils.GetBody[*prot.User](c, func() *prot.User {
		return &prot.User{}
	})

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	currentUser, err := sc.svc.GetByID(c, int(id))

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_record_update")), "messages.no_record_update")
		return
	}

	req.Id = id
	req.RoleId = int64(currentUser.RoleId)

	user, err := sc.svc.UpdateProfile(c, req)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	userResource := resources.NewUserResource()
	formattedUser := userResource.FormatUser(user)

	utils.Respond(c, formattedUser, nil, "")
}

func (sc *UserController) UsernameExits(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.username_required")), "messages.username_required")
		return
	}

	exists, err := sc.svc.UsernameExits(c, username)
	if err != nil {
		utils.Respond(c, nil, err, "messages.data_existed")
		return
	}

	response := &prot.UsernameExits{
		Username: username,
		Exits:    exists,
	}

	utils.Respond(c, response, nil, "")
}

func (sc *UserController) GetPermissions(c *gin.Context) {
	id := utils.GetCurrentRoleId(c)

	roleRepo := repositories.NewRoleRepository()
	roleService := services.NewRoleService(roleRepo)

	permissions, err := roleService.GetPermissions(c, id)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_records_found")), "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, permissions, err, "")
}

func (sc *UserController) CurrentClass(c *gin.Context) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId != models.StudentRoleId && roleId != models.TeacherRoleId {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.role_invalid")), "messages.role_invalid")
		return
	}

	data, err := sc.svc.CurrentClass(c)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	userResource := resources.NewClassResource()
	formattedUser := userResource.FormatClassByUserReponse(data)

	utils.Respond(c, formattedUser, err, "")
}

func (cc *UserController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func() {
		url, err := cc.svc.Export(cCp)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (cc *UserController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := cc.svc.Import(cCp, file)
		if err != nil {
			config.Log.Error("User import failed", "error", err)
		} else {
			config.Log.Info("User import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "messages.import_complete")
}

func (uc *UserController) UserActivated(c *gin.Context) {
	users, count, err := uc.svc.UserActivated(c)

	userResource := resources.NewUserResource()
	userFormat := userResource.FormatUsers(users)

	utils.Respond(c, &prot.UsersResponse{
		Users:      userFormat,
		TotalCount: count,
	}, err, "")
}

func (sc *UserController) ResetPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	err, message := sc.svc.ResetPassword(c, id)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize(message)), message, http.StatusNotFound)
		return
	}

	utils.Respond(c, nil, err, "")
}

func (uc *UserController) RespondList(c *gin.Context, items []models.User, totalCount int64, err error) {
	var userPtrs []*models.User
	for i := range items {
		userPtrs = append(userPtrs, &items[i])
	}

	userResource := resources.NewUserResource()

	if programIDStr := c.Query("program_id"); programIDStr != "" {
		if programID, err := strconv.ParseInt(programIDStr, 10, 64); err == nil {
			programRepo := repositories.NewProgramRepository()
			failedUserIds, _ := programRepo.GetFailedUserIdsById(programID)

			if len(failedUserIds) > 0 {
				if impl, ok := userResource.(*resources.UserResourceImpl); ok {
					impl.FailedUserIds = failedUserIds
				}
			}
		}
	}

	usersResponse := userResource.FormatUsers(userPtrs)

	list := &prot.UsersResponse{
		Users:      usersResponse,
		TotalCount: int64(totalCount),
	}

	utils.Respond(c, list, err, "")
}
