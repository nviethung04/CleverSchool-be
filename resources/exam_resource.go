package resources

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"time"
)

type ExamResource interface {
	FormatExam(exam *models.Exam) *prot.Exam
	FormatExams(exams []*models.Exam) []*prot.Exam
	FormatModelExam(exam *prot.ExamRequest) *models.Exam
}

type ExamResourceImpl struct {
	CourseId int64
}

func NewExamResource() ExamResource {
	return &ExamResourceImpl{
		CourseId: 0,
	}
}

func (r *ExamResourceImpl) FormatExam(exam *models.Exam) *prot.Exam {
	if exam == nil {
		return nil
	}

	var isAssigned bool

	for _, ref := range exam.ExamRefLessons {
		if ref.ExamId == exam.ID && ref.CourseId == r.CourseId {
			isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
		}
		if isAssigned {
			break
		}
	}

	questionFiles := make([]*prot.ExamFile, 0, len(exam.FileInfos))

	for _, fileInfo := range exam.FileInfos {
		questionFiles = append(questionFiles, &prot.ExamFile{
			Type: fileInfo.Type,
			Url:  utils.StaticURL(fileInfo.Path, models.Storage),
		})
	}

	return &prot.Exam{
		Id:               exam.ID,
		Name:             exam.Name,
		Status:           int32(exam.Status),
		TimeLimit:        exam.TimeLimit,
		MaxScore:         float32(exam.MaxScore),
		Description:      exam.Description,
		CoverImage:       utils.StaticURL(exam.CoverImageInfo.Path, models.Storage),
		CreatedAt:        exam.CreatedAt.Unix(),
		CreatedBy:        exam.CreatedBy,
		UpdatedAt:        exam.UpdatedAt.Unix(),
		UpdatedBy:        exam.UpdatedBy,
		Deadline:         exam.Deadline.Unix(),
		IsAssigned:       isAssigned,
		TotalQuestions:   int32(exam.TotalQuestions),
		ObjectTitle:      exam.ObjectTitle,
		IsRandomQuestion: exam.IsRandomQuestion,
		QuestionForm:     exam.QuestionForm,
		QuestionFiles:    questionFiles,
		Type:             exam.Type,
	}
}

func (r *ExamResourceImpl) FormatExams(exams []*models.Exam) []*prot.Exam {
	result := make([]*prot.Exam, 0, len(exams))
	for _, e := range exams {
		if formatted := r.FormatExam(e); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ExamResourceImpl) FormatModelExam(exam *prot.ExamRequest) *models.Exam {
	if exam == nil {
		return nil
	}

	coverImageUrl := utils.StripDomain(exam.CoverImage, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	coverImageInfo := mediaRepo.GetMediaInfo(coverImageUrl, models.Storage)

	deadlineTime := time.Unix(exam.Deadline, 0)

	fileInfos := make([]models.MediaDetail, 0, len(exam.QuestionFiles))

	for _, file := range exam.QuestionFiles {
		fileUrl := utils.StripDomain(file.Url, models.Storage)
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		fileInfos = append(fileInfos, models.MediaDetail{
			Path: utils.StripDomain(fileInfo.Path, models.Storage),
			Id:   fileInfo.Id,
			Type: file.Type,
			Disk: fileInfo.Disk,
		})
	}

	questionForm := exam.QuestionForm

	allowed := map[string]struct{}{
		"question": {},
		"write":    {},
	}

	if _, ok := allowed[questionForm]; !ok {
		questionForm = models.QuestionFormQuestionType
	}

	isVtg := config.LoadConfig().IsVtg
	examType := ""

	if isVtg {
		if exam.Type != models.ExamTypeFrequent && exam.Type != models.ExamTypeEvaluate {
			examType = models.ExamTypeFrequent
		} else {
			examType = exam.Type
		}
	} else {
		examType = models.ExamTypeExam
	}

	return &models.Exam{
		ID:               exam.Id,
		Name:             exam.Name,
		Status:           int16(exam.Status),
		TimeLimit:        exam.TimeLimit,
		MaxScore:         float64(exam.MaxScore),
		Description:      exam.Description,
		CoverImageInfo:   coverImageInfo,
		Deadline:         deadlineTime,
		TotalQuestions:   exam.TotalQuestions,
		ObjectTitle:      exam.ObjectTitle,
		IsRandomQuestion: exam.IsRandomQuestion,
		FileInfos:        fileInfos,
		QuestionForm:     questionForm,
		Type:             examType,
	}
}
