package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
)

type SkillResource interface {
	FormatSkill(skill *models.Skill) *prot.Skill
	FormatSkills(skills []*models.Skill) []*prot.Skill
	FormatModelSkill(skill *prot.SkillRequest) *models.Skill
}

type SkillResourceImpl struct{}

func NewSkillResource() SkillResource {
	return &SkillResourceImpl{}
}

func (r *SkillResourceImpl) FormatSkill(skill *models.Skill) *prot.Skill {
	if skill == nil {
		return nil
	}

	return &prot.Skill{
		Id:          int64(skill.ID),
		Name:        skill.Name,
		Type:        skill.Type,
		ParentId:    utils.Int64OrZero(skill.ParentID),
		Description: skill.Description,
		ImageUrl:    utils.StaticURL(skill.ImageInfo.Path, models.Storage),
		Status:      skill.Status,
		CreatedAt:   skill.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   skill.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SkillResourceImpl) FormatSkills(skills []*models.Skill) []*prot.Skill {
	result := make([]*prot.Skill, 0, len(skills))
	for _, t := range skills {
		if formatted := r.FormatSkill(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *SkillResourceImpl) FormatModelSkill(skill *prot.SkillRequest) *models.Skill {
	if skill == nil {
		return nil
	}

	imageUrl := utils.StripDomain(skill.ImageUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(imageUrl, models.Storage)

	return &models.Skill{
		ID:          int64(skill.Id),
		Name:        skill.Name,
		Type:        skill.Type,
		ParentID:    utils.Int64PtrOrNil(skill.ParentId),
		Description: skill.Description,
		ImageInfo:   imageInfo,
		Status:      skill.Status,
	}
}
