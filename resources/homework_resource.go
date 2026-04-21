package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
)

type HomeworkResource interface {
	FormatHomework(homework *models.Homework) *prot.Homework
	FormatHomeworks(homeworks []*models.Homework) []*prot.Homework
	FormatModelHomework(homework *prot.HomeworkRequest) *models.Homework
}

type HomeworkResourceImpl struct{
	CourseId int64
}

func NewHomeworkResource() HomeworkResource {
	return &HomeworkResourceImpl{
		CourseId: 0,
	}
}

func (r *HomeworkResourceImpl) FormatHomework(homework *models.Homework) *prot.Homework {
	if homework == nil {
		return nil
	}

	var isAssigned bool

	for _, ref := range homework.HomeworkRefLessons {
		if ref.HomeworkId == homework.ID && ref.CourseId == r.CourseId {
			isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
		}
		if isAssigned {
			break
		}
	}

	questionFiles := make([]*prot.HomeworkFile, 0, len(homework.FileInfos))

	for _, fileInfo := range homework.FileInfos {
		questionFiles = append(questionFiles, &prot.HomeworkFile{
			Type: fileInfo.Type,
			Url:  utils.StaticURL(fileInfo.Path, models.Storage),
		})
	}

	return &prot.Homework{
		Id:             homework.ID,
		LessonId:       0,
		Name:           homework.Name,
		Status:         int32(homework.Status),
		Description:    homework.Description,
		ObjectTitle:	homework.ObjectTitle,
		MaxScore:       float32(homework.MaxScore),
		CoverImage:     utils.StaticURL(homework.CoverImageInfo.Path, models.Storage),
		CreatedAt:      homework.CreatedAt.Unix(),
		CreatedBy:      homework.CreatedBy,
		UpdatedAt:      homework.UpdatedAt.Unix(),
		UpdatedBy:      homework.UpdatedBy,
		IsAssigned:     isAssigned,
		TotalQuestions: homework.TotalQuestions,
		IsRandomQuestion: homework.IsRandomQuestion,
		QuestionForm: homework.QuestionForm,
		QuestionFiles: questionFiles,
	}
}

func (r *HomeworkResourceImpl) FormatHomeworks(exams []*models.Homework) []*prot.Homework {
	result := make([]*prot.Homework, 0, len(exams))
	for _, e := range exams {
		if formatted := r.FormatHomework(e); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *HomeworkResourceImpl) FormatModelHomework(homework *prot.HomeworkRequest) *models.Homework {
	if homework == nil {
		return nil
	}

	mediaRepo := repositories.NewMediaRepository()

	coverImageUrl := utils.StripDomain(homework.CoverImage, models.Storage)
	coverImageInfo := mediaRepo.GetMediaInfo(coverImageUrl, models.Storage)

	fileInfos := make([]models.MediaDetail, 0, len(homework.QuestionFiles))
	for _, file := range homework.QuestionFiles {
		fileUrl := utils.StripDomain(file.Url, models.Storage)
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		fileInfos = append(fileInfos, models.MediaDetail{
			Path: utils.StripDomain(fileInfo.Path, models.Storage),
			Id:   fileInfo.Id,
			Type: file.Type,
			Disk: fileInfo.Disk,
		})
	}

	questionForm := homework.QuestionForm

	allowed := map[string]struct{}{
		"question": {},
		"write":    {},
	}

	if _, ok := allowed[questionForm]; !ok {
		questionForm = models.QuestionFormQuestionType
	}

	return &models.Homework{
		ID:             homework.Id,
		Name:           homework.Name,
		ObjectTitle:	homework.ObjectTitle,
		Status:         int16(homework.Status),
		MaxScore:       float64(homework.MaxScore),
		Description:    homework.Description,
		CoverImageInfo: coverImageInfo,
		// IsAssigned:     &homework.IsAssigned,
		TotalQuestions: homework.TotalQuestions,
		IsRandomQuestion: homework.IsRandomQuestion,
		FileInfos: fileInfos,
		QuestionForm: questionForm,
	}
}
