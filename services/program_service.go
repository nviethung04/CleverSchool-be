package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProgramService interface {
	GetAll(c *gin.Context) ([]models.Program, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Program, error)
	Create(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error)
	Update(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Program, error)
	SortChapters(c *gin.Context, id int, req *prot.ProgramChapterSort) error
	Export(c *gin.Context, id int) (string, error)
	Import(c *gin.Context, file *multipart.FileHeader) error
}

type programService struct {
	repo repositories.ProgramRepository
}

func NewProgramService(repo repositories.ProgramRepository) ProgramService {
	return &programService{repo: repo}
}

func (s *programService) GetAll(c *gin.Context) ([]models.Program, int64, error) {
	allowedFilters := []string{}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	filter, _ = s.ApplyFilter(c, filter)

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Courses",
		"Subjects",
	})

	programs, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return programs, rows, nil
}

func (s *programService) GetByID(c *gin.Context, id int) (*prot.Program, error) {
	s.repo.SetPreload([]string{
		"Courses",
		"Subjects",
		"Chapters",
		"Chapters.Lessons",
		"Chapters.Headings",
	})
	s.repo.SetContext(c)
	program, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	programResource := resources.NewProgramResource()
	formattedProgram := programResource.FormatProgram(program)

	return formattedProgram, nil
}

func (s *programService) Create(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error) {
	programResource := resources.NewProgramResource()
	program := programResource.FormatModelProgram(req)

	s.repo.SetContext(c)

	err := s.repo.Create(program)
	if err != nil {
		return nil, err
	}

	s.StoreProgramSubjects(int64(program.ID), req)

	id := int(program.ID)
	s.repo.SetPreload([]string{
		"Courses",
		"Subjects",
		"Chapters",
		"Chapters.Lessons",
		"Chapters.Headings",
	})
	newProgram, _ := s.repo.FindNewByID(id)

	return newProgram, nil
}

func (s *programService) Update(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error) {
	programResource := resources.NewProgramResource()
	program := programResource.FormatModelProgram(req)

	s.repo.SetContext(c)

	err := s.repo.Update(program)
	if err != nil {
		return nil, err
	}

	s.StoreProgramSubjects(int64(program.ID), req)

	s.repo.ChapterUpdateSortPosition(program.ID)

	id := int(program.ID)
	s.repo.SetPreload([]string{
		"Courses",
		"Subjects",
		"Chapters",
		"Chapters.Lessons",
		"Chapters.Headings",
	})
	updatedProgram, _ := s.repo.FindNewByID(id)

	return updatedProgram, nil
}

func (s *programService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)

	return s.repo.Delete(id)
}

func (s *programService) Restore(c *gin.Context, id int) (*models.Program, error) {
	s.repo.SetContext(c)
	program, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return program, nil
}

func (s *programService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	var programIDs []int64
	hasFilter := false

	if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
		if schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64); err == nil {
			ids := s.repo.GetIdsBySchoolId(schoolID)

			programIDs = ids
			hasFilter = true
		}
	}

	var (
		facultyID int64
		subjectID int64
	)

	if facultyIDStr := c.Query("faculty_id"); facultyIDStr != "" {
		facultyID, _ = strconv.ParseInt(facultyIDStr, 10, 64)
	}

	if subjectIDStr := c.Query("subject_id"); subjectIDStr != "" {
		subjectID, _ = strconv.ParseInt(subjectIDStr, 10, 64)
	}

	if facultyID > 0 || subjectID > 0 {
		ids := s.repo.GetIdsBySubjectAndFaculty(subjectID, facultyID)

		if hasFilter {
			programIDs = intersectInt64(programIDs, ids)
		} else {
			programIDs = ids
			hasFilter = true
		}
	}

	if hasFilter {
		if len(programIDs) == 0 {
			filter["programs.id"] = "in:-1"
		} else {
			idStrs := make([]string, len(programIDs))
			for i, id := range programIDs {
				idStrs[i] = strconv.FormatInt(id, 10)
			}
			filter["programs.id"] = "in:" + strings.Join(idStrs, ",")
		}
	}

	return filter, nil
}

func (s *programService) StoreProgramSubjects(programID int64, program *prot.ProgramRequest) {
	programSubjects := make([]models.ProgramRefSubject, 0)

	for _, subject := range program.Subjects {
		programSubjects = append(programSubjects, models.ProgramRefSubject{
			ProgramId: programID,
			SubjectId: int64(subject.Id),
		})
	}

	if len(programSubjects) == 0 {
		if program.SubjectId != 0 {
			programSubjects = append(programSubjects, models.ProgramRefSubject{
				ProgramId: programID,
				SubjectId: int64(program.SubjectId),
			})
		}
	}

	if len(programSubjects) == 0 {
		s.repo.DeleteProgramRefSubjectsByProgramID(programID)
	} else {
		subjectIDs := make(map[uint]bool)

		for _, ps := range programSubjects {
			subjectIDs[uint(ps.SubjectId)] = true
		}

		existingProgramSubjects, _ := s.repo.GetProgramRefSubjectsByProgramID(programID)

		for _, eps := range existingProgramSubjects {
			if !subjectIDs[uint(eps.SubjectId)] {
				s.repo.DeleteProgramRefSubjectsByProgramAndSubjectID(programID, int64(eps.SubjectId))
			}
		}

		for _, ps := range programSubjects {
			s.repo.UpdateOrCreateProgramRefSubject(ps)
		}
	}
}

func intersectInt64(a, b []int64) []int64 {
	if len(a) == 0 || len(b) == 0 {
		return []int64{}
	}

	set := make(map[int64]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}

	var result []int64
	for _, v := range b {
		if _, ok := set[v]; ok {
			result = append(result, v)
		}
	}
	return result
}

func (s *programService) SortChapters(c *gin.Context, id int, req *prot.ProgramChapterSort) error {
	repo := repositories.NewChapterRepository()
	for index, chapter := range req.Chapters {
		repo.UpdatePositionById(id, int(chapter.Id), index)
	}
	return nil
}
