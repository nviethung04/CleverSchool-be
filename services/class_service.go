package services

import (
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"errors"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClassService interface {
	GetAll(c *gin.Context) ([]models.Class, int64, error)
	GetByID(c *gin.Context, id int) (*prot.ClassResponse, error)
	Create(c *gin.Context, req *prot.ClassRequest) (*models.Class, error)
	Update(c *gin.Context, req *prot.ClassRequest) (*models.Class, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Class, error)
	GetUsers(c *gin.Context, id int64) ([]models.User, error)
	StoreUsers(c *gin.Context, id int64) ([]models.User, error)
	AddUsers(c *gin.Context, id int64) ([]models.User, error)
	Export(c *gin.Context) (string, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
}

type classService struct {
	repo repositories.ClassRepository
}

func NewClassService(repo repositories.ClassRepository) ClassService {
	return &classService{repo}
}

func (s *classService) GetAll(c *gin.Context) ([]models.Class, int64, error) {
	allowedFilters := []string{"status", "school_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Grade",
	})

	classes, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return classes, rows, nil
}

func (s *classService) GetByID(c *gin.Context, id int) (*prot.ClassResponse, error) {
	s.repo.SetPreload([]string{
		"Grade",
	})

	class, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	classResource := resources.NewClassResource()
	formattedClass := classResource.FormatClass(class)

	return formattedClass, nil
}

func (s *classService) Create(c *gin.Context, req *prot.ClassRequest) (*models.Class, error) {
	classResource := resources.NewClassResource()
	class := classResource.FormatModelClass(req)

	s.repo.SetContext(c)

	err := s.repo.Create(class)
	if err != nil {
		return nil, err
	}

	id := int(class.ID)
	s.repo.UpdateCurrentStudentToClass(int64(id))
	s.repo.SetPreload([]string{
		"Grade",
	})
	newClass, _ := s.repo.FindNewByID(id)

	return newClass, nil
}

func (s *classService) Update(c *gin.Context, req *prot.ClassRequest) (*models.Class, error) {
	classResource := resources.NewClassResource()
	class := classResource.FormatModelClass(req)

	s.repo.SetContext(c)

	err := s.repo.Update(class)
	if err != nil {
		return nil, err
	}

	id := int(class.ID)
	s.repo.UpdateCurrentStudentToClass(int64(id))
	s.repo.SetPreload([]string{
		"Grade",
	})
	updateClass, _ := s.repo.FindNewByID(id)

	return updateClass, nil
}

func (s *classService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	s.repo.UpdateCurrentStudentToClass(int64(id))
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *classService) Restore(c *gin.Context, id int) (*models.Class, error) {
	s.repo.SetContext(c)
	class, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return class, nil
}

func (s *classService) GetUsers(c *gin.Context, id int64) ([]models.User, error) {
	role := c.Query("role")
	roleIdStr := c.Query("role_id")

	roleRepo := repositories.NewRoleRepository()

	var roleModel *models.Role
	var err error

	if role != "" {
		roleModel, err = roleRepo.FindByName(role)
	} else if roleIdStr != "" {
		roleId, err := strconv.Atoi(roleIdStr)
		if err != nil {
			return nil, err
		}
		roleModel, err = roleRepo.FindByID(roleId)
	}

	students, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, err
	}

	return students, nil
}

func (s *classService) StoreUsers(c *gin.Context, id int64) ([]models.User, error) {
	var req prot.UserClassRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	roleRepo := repositories.NewRoleRepository()
	roleModel, err := roleRepo.FindByID(int(req.RoleId))

	if err != nil {
		return nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	userIds := make([]int64, 0, len(req.Users))
	for _, u := range req.Users {
		userIds = append(userIds, u.Id)
	}

	// class, err := s.repo.FindByID(int(id))

	// if err != nil {
	// 	return nil, errors.New(i18n.Localize("messages.data_invalid"))
	// }

	// if len(userIds) > int(class.MaxStudents) {
	// 	return nil, errors.New(i18n.Localize("messages.max_student_invalid", map[string]interface{}{"Count": class.MaxStudents}))
	// }

	if err := s.repo.ReplaceUserClass(id, userIds, roleModel.ID); err != nil {
		return nil, err
	}

	newStudents, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, err
	}

	return newStudents, nil
}

func (s *classService) AddUsers(c *gin.Context, id int64) ([]models.User, error) {
	var req prot.UserClassRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	roleRepo := repositories.NewRoleRepository()
	roleModel, err := roleRepo.FindByID(int(req.RoleId))

	if err != nil {
		return nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	userIds := make([]int64, 0, len(req.Users))
	for _, u := range req.Users {
		userIds = append(userIds, u.Id)
	}

	// class, err := s.repo.FindByID(int(id))

	// if err != nil {
	// 	return nil, errors.New(i18n.Localize("messages.data_invalid"))
	// }

	// if len(userIds)+int(class.CurrentStudents) > int(class.MaxStudents) {
	// 	return nil, errors.New(i18n.Localize("messages.max_student_invalid", map[string]interface{}{"Count": class.MaxStudents}))
	// }

	if err := s.repo.AddUserClass(id, userIds, roleModel.ID); err != nil {
		return nil, err
	}

	newStudents, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, err
	}

	return newStudents, nil
}
