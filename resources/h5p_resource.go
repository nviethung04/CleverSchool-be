package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"encoding/json"

	"google.golang.org/protobuf/types/known/structpb"
)

type H5PResource interface {
	FormatH5P(h5p *models.H5pContent) *prot.H5P
	FormatH5Ps(h5ps []*models.H5pContent) []*prot.H5P
}

type H5PResourceImpl struct{}

func NewH5PResource() H5PResource {
	return &H5PResourceImpl{}
}

func (r *H5PResourceImpl) FormatH5P(h5p *models.H5pContent) *prot.H5P {
	if h5p == nil {
		return nil
	}

	var parametersStruct, metadataStruct *structpb.Struct
	var err error

	if len(h5p.Parameters) > 0 {
		var paramsMap map[string]interface{}
		if err = json.Unmarshal(h5p.Parameters, &paramsMap); err == nil {
			parametersStruct, _ = structpb.NewStruct(paramsMap)
		}
	}
	return &prot.H5P{
		Id:          int64(h5p.ID),
		ContentId:   h5p.ContentID,
		Title:        h5p.Title,
		Library:        h5p.Library,
		Parameters:   parametersStruct,
		Metadata:   metadataStruct,
		CreatedAt:   h5p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   h5p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *H5PResourceImpl) FormatH5Ps(h5ps []*models.H5pContent) []*prot.H5P {
	result := make([]*prot.H5P, 0, len(h5ps))
	for _, t := range h5ps {
		if formatted := r.FormatH5P(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}
