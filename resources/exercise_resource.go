package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"time"
)

type ExerciseResource interface {
	FormatExercise(exercise *models.Exercise) *prot.Exercise
	FormatExercises(exercises []*models.Exercise) []*prot.Exercise
	FormatModelExercise(exercise *prot.ExerciseRequest) *models.Exercise
}

type ExerciseResourceImpl struct{
	CourseId int64
}

func NewExerciseResource() ExerciseResource {
	return &ExerciseResourceImpl{
		CourseId: 0,
	}
}

func (r *ExerciseResourceImpl) FormatExercise(exercise *models.Exercise) *prot.Exercise {
	if exercise == nil {
		return nil
	}

	var isAssigned bool

	for _, ref := range exercise.ExerciseRefLessons {
		if ref.ExerciseId == exercise.ID && ref.CourseId == r.CourseId {
			isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
		}
		if isAssigned {
			break
		}
	}

	questionFiles := make([]*prot.ExerciseFile, 0, len(exercise.FileInfos))

	for _, fileInfo := range exercise.FileInfos {
		questionFiles = append(questionFiles, &prot.ExerciseFile{
			Type: fileInfo.Type,
			Url:  utils.StaticURL(fileInfo.Path, models.Storage),
		})
	}

	return &prot.Exercise{
		Id:             exercise.ID,
		Name:           exercise.Name,
		Status:         int32(exercise.Status),
		TimeLimit:      exercise.TimeLimit,
		MaxScore:       float32(exercise.MaxScore),
		Description:    exercise.Description,
		CoverImage:     utils.StaticURL(exercise.CoverImageInfo.Path, models.Storage),
		CreatedAt:      exercise.CreatedAt.Unix(),
		CreatedBy:      exercise.CreatedBy,
		UpdatedAt:      exercise.UpdatedAt.Unix(),
		UpdatedBy:      exercise.UpdatedBy,
		Deadline:       exercise.Deadline.Unix(),
		IsAssigned:     isAssigned,
		TotalQuestions: int32(exercise.TotalQuestions),
		ObjectTitle:    exercise.ObjectTitle,
		IsRandomQuestion: exercise.IsRandomQuestion,
		QuestionForm: exercise.QuestionForm,
		QuestionFiles: questionFiles,
	}
}

func (r *ExerciseResourceImpl) FormatExercises(exercises []*models.Exercise) []*prot.Exercise {
	result := make([]*prot.Exercise, 0, len(exercises))
	for _, e := range exercises {
		if formatted := r.FormatExercise(e); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ExerciseResourceImpl) FormatModelExercise(exercise *prot.ExerciseRequest) *models.Exercise {
	if exercise == nil {
		return nil
	}
	mediaRepo := repositories.NewMediaRepository()

	deadlineTime := time.Unix(exercise.Deadline, 0)

	fileInfos := make([]models.MediaDetail, 0, len(exercise.QuestionFiles))
	for _, file := range exercise.QuestionFiles {
		fileUrl := utils.StripDomain(file.Url, models.Storage)
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		fileInfos = append(fileInfos, models.MediaDetail{
			Path: utils.StripDomain(fileInfo.Path, models.Storage),
			Id:   fileInfo.Id,
			Type: file.Type,
			Disk: fileInfo.Disk,
		})
	}

	questionForm := exercise.QuestionForm

	allowed := map[string]struct{}{
		"question": {},
		"write":    {},
	}

	if _, ok := allowed[questionForm]; !ok {
		questionForm = models.QuestionFormQuestionType
	}

	return &models.Exercise{
		ID:             exercise.Id,
		Name:           exercise.Name,
		Status:         int16(exercise.Status),
		TimeLimit:      exercise.TimeLimit,
		MaxScore:       float64(exercise.MaxScore),
		Description:    exercise.Description,
		Deadline:       deadlineTime,
		TotalQuestions: int32(exercise.TotalQuestions),
		ObjectTitle:    exercise.ObjectTitle,
		IsRandomQuestion: exercise.IsRandomQuestion,
		FileInfos: fileInfos,
		QuestionForm: questionForm,
	}
}

