package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"fmt"
	"time"
)

type ContestResource interface {
	FormatContest(contest *models.Contest) *prot.Contest
	FormatContests(contests []*models.Contest) []*prot.Contest
	FormatModelContest(contest *prot.ContestRequest) *models.Contest
}

type ContestResourceImpl struct{}

func NewContestResource() ContestResource {
	return &ContestResourceImpl{}
}

func (r *ContestResourceImpl) FormatContest(contest *models.Contest) *prot.Contest {
	if contest == nil {
		return nil
	}

	protoContest := &prot.Contest{
		Id:          contest.ID,
		Name:        contest.Name,
		Description: contest.Description,
		Status:      int32(contest.Status),
		CreatedAt:   contest.CreatedAt.Unix(),
		CreatedBy:   contest.CreatedBy,
		UpdatedAt:   contest.UpdatedAt.Unix(),
		UpdatedBy:   contest.UpdatedBy,
		DeletedAt:   contest.DeletedAt.Time.Unix(),
		DeletedBy:   contest.DeletedBy,
		Image:       utils.StaticURL(contest.ImageInfo.Path, models.Storage),
	}

	if contest.StartTime != nil {
		protoContest.StartTime = contest.StartTime.Unix()
	}
	if contest.EndTime != nil {
		protoContest.EndTime = contest.EndTime.Unix()
	}

	// Format creator information
	if contest.Creator != nil {
		fmt.Printf("🔍 DEBUG: Creator found - ID: %d, Name: %s, Email: %s, Username: %s\n",
			contest.Creator.ID, contest.Creator.Name, contest.Creator.Email, contest.Creator.Username)
		protoContest.Creator = &prot.ContestCreator{
			Id:       contest.Creator.ID,
			Name:     contest.Creator.Name,
			Email:    contest.Creator.Email,
			Username: contest.Creator.Username,
		}
	} else {
		fmt.Printf("🔍 DEBUG: Creator is nil for contest ID: %d, CreatedBy: %d\n", contest.ID, contest.CreatedBy)
	}

	// Format contest rounds
	for _, round := range contest.ContestRounds {
		questionFiles := make([]*prot.ContestRoundFile, 0, len(round.FileInfos))

		for _, fileInfo := range round.FileInfos {
			questionFiles = append(questionFiles, &prot.ContestRoundFile{
				Type: fileInfo.Type,
				Url:  utils.StaticURL(fileInfo.Path, models.Storage),
			})
		}

		protoRound := &prot.ContestRound{
			Id:           round.ID,
			ContestId:    round.ContestId,
			Name:         round.Name,
			Description:  round.Description,
			SortPosition: round.SortPosition,
			JoinLevel:    round.JoinLevel,
			CreatedAt:    round.CreatedAt.Unix(),
			CreatedBy:    round.CreatedBy,
			UpdatedAt:    round.UpdatedAt.Unix(),
			UpdatedBy:    round.UpdatedBy,
			DeletedAt:    round.DeletedAt.Time.Unix(),
			DeletedBy:    round.DeletedBy,
			QuestionForm: round.QuestionForm,
			QuestionFiles: questionFiles,
		}

		if round.StartTime != nil {
			protoRound.StartTime = round.StartTime.Unix()
		}
		if round.EndTime != nil {
			protoRound.EndTime = round.EndTime.Unix()
		}

		protoContest.ContestRounds = append(protoContest.ContestRounds, protoRound)
	}

	return protoContest
}

func (r *ContestResourceImpl) FormatContests(contests []*models.Contest) []*prot.Contest {
	result := make([]*prot.Contest, 0, len(contests))
	for _, c := range contests {
		if formatted := r.FormatContest(c); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ContestResourceImpl) FormatModelContest(contest *prot.ContestRequest) *models.Contest {
	if contest == nil {
		return nil
	}

	modelContest := &models.Contest{
		ID:          contest.Id,
		Name:        contest.Name,
		Description: contest.Description,
		Status:      int16(contest.Status),
		CreatedBy:   contest.CreatedBy,
		UpdatedBy:   contest.UpdatedBy,
		DeletedBy:   contest.DeletedBy,
	}

	if contest.StartTime > 0 {
		startTime := time.Unix(contest.StartTime, 0)
		modelContest.StartTime = &startTime
	}
	if contest.EndTime > 0 {
		endTime := time.Unix(contest.EndTime, 0)
		modelContest.EndTime = &endTime
	}

	// Handle image info
	if contest.ImageInfo != nil {
		modelContest.ImageInfo = models.MediaInfo{
			Id:   contest.ImageInfo.Id,
			Path: contest.ImageInfo.Path,
			Disk: contest.ImageInfo.Disk,
		}
	}

	return modelContest
}

type ContestRoundResource interface {
	FormatContestRound(round *models.ContestRound) *prot.ContestRound
	FormatContestRounds(rounds []*models.ContestRound) []*prot.ContestRound
	FormatModelContestRound(round *prot.ContestRoundRequest) *models.ContestRound
}

type ContestRoundResourceImpl struct{}

func NewContestRoundResource() ContestRoundResource {
	return &ContestRoundResourceImpl{}
}

func (r *ContestRoundResourceImpl) FormatContestRound(round *models.ContestRound) *prot.ContestRound {
	if round == nil {
		return nil
	}

	protoRound := &prot.ContestRound{
		Id:           round.ID,
		ContestId:    round.ContestId,
		Name:         round.Name,
		Description:  round.Description,
		SortPosition: round.SortPosition,
		JoinLevel:    round.JoinLevel,
		CreatedAt:    round.CreatedAt.Unix(),
		CreatedBy:    round.CreatedBy,
		UpdatedAt:    round.UpdatedAt.Unix(),
		UpdatedBy:    round.UpdatedBy,
		DeletedAt:    round.DeletedAt.Time.Unix(),
		DeletedBy:    round.DeletedBy,
	}

	if round.StartTime != nil {
		protoRound.StartTime = round.StartTime.Unix()
	}
	if round.EndTime != nil {
		protoRound.EndTime = round.EndTime.Unix()
	}

	return protoRound
}

func (r *ContestRoundResourceImpl) FormatContestRounds(rounds []*models.ContestRound) []*prot.ContestRound {
	result := make([]*prot.ContestRound, 0, len(rounds))
	for _, round := range rounds {
		if formatted := r.FormatContestRound(round); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ContestRoundResourceImpl) FormatModelContestRound(round *prot.ContestRoundRequest) *models.ContestRound {
	if round == nil {
		return nil
	}
	mediaRepo := repositories.NewMediaRepository()

	fileInfos := make([]models.MediaDetail, 0, len(round.QuestionFiles))
	for _, file := range round.QuestionFiles {
		fileUrl := utils.StripDomain(file.Url, models.Storage)
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		fileInfos = append(fileInfos, models.MediaDetail{
			Path: utils.StripDomain(fileInfo.Path, models.Storage),
			Id:   fileInfo.Id,
			Type: file.Type,
			Disk: fileInfo.Disk,
		})
	}

	questionForm := round.QuestionForm

	allowed := map[string]struct{}{
		"question": {},
		"write":    {},
	}

	if _, ok := allowed[questionForm]; !ok {
		questionForm = models.QuestionFormQuestionType
	}

	modelRound := &models.ContestRound{
		ID:           round.Id,
		ContestId:    round.ContestId,
		Name:         round.Name,
		Description:  round.Description,
		SortPosition: round.SortPosition,
		JoinLevel:    round.JoinLevel,
		CreatedBy:    round.CreatedBy,
		UpdatedBy:    round.UpdatedBy,
		DeletedBy:    round.DeletedBy,
		FileInfos: fileInfos,
		QuestionForm: questionForm,
	}

	if round.StartTime > 0 {
		startTime := time.Unix(round.StartTime, 0)
		modelRound.StartTime = &startTime
	}
	if round.EndTime > 0 {
		endTime := time.Unix(round.EndTime, 0)
		modelRound.EndTime = &endTime
	}

	return modelRound
}

