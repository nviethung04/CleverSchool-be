package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
)

type TagResource interface {
	FormatTag(tag *models.Tag) *prot.Tag
	FormatTags(tags []*models.Tag) []*prot.Tag
	FormatModelTag(tag *prot.TagRequest) *models.Tag
}

type TagResourceImpl struct{}

func NewTagResource() TagResource {
	return &TagResourceImpl{}
}

func (r *TagResourceImpl) FormatTag(tag *models.Tag) *prot.Tag {
	if tag == nil {
		return nil
	}

	return &prot.Tag{
		Id:          int64(tag.ID),
		Name:        tag.Name,
		Type:        tag.Type,
		Description: tag.Description,
		ImageUrl:    utils.StaticURL(tag.ImageInfo.Path, models.Storage),
		Status:      tag.Status,
		CreatedAt:   tag.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   tag.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *TagResourceImpl) FormatTags(tags []*models.Tag) []*prot.Tag {
	result := make([]*prot.Tag, 0, len(tags))
	for _, t := range tags {
		if formatted := r.FormatTag(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *TagResourceImpl) FormatModelTag(tag *prot.TagRequest) *models.Tag {
	if tag == nil {
		return nil
	}

	imageUrl := utils.StripDomain(tag.ImageUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(imageUrl, models.Storage)

	return &models.Tag{
		ID:          int64(tag.Id),
		Name:        tag.Name,
		Type:        tag.Type,
		Description: tag.Description,
		ImageInfo:   imageInfo,
		Status:      tag.Status,
	}
}

