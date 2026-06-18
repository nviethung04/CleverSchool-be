package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type SettingResource interface {
	FormatSetting(setting *models.Setting) *prot.Setting
	FormatSettings(settings []*models.Setting) []*prot.Setting
	FormatModelSetting(req *prot.SettingRequest) *models.Setting
}

type SettingResourceImpl struct{}

func NewSettingResource() SettingResource {
	return &SettingResourceImpl{}
}

func (r *SettingResourceImpl) FormatSetting(setting *models.Setting) *prot.Setting {
	if setting == nil {
		return nil
	}

	values := make([]*prot.SettingKVPair, 0, len(setting.Values))
	for _, pair := range setting.Values {
		values = append(values, &prot.SettingKVPair{
			Key:   pair.Key,
			Value: pair.Value,
		})
	}

	return &prot.Setting{
		Id:        setting.ID,
		Key:       setting.Key,
		Name:      setting.Name,
		Values:    values,
		IsActive:  setting.IsActive,
		CreatedBy: setting.CreatedBy,
		UpdatedBy: setting.UpdatedBy,
		CreatedAt: setting.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: setting.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SettingResourceImpl) FormatSettings(settings []*models.Setting) []*prot.Setting {
	result := make([]*prot.Setting, 0, len(settings))
	for _, item := range settings {
		if formatted := r.FormatSetting(item); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *SettingResourceImpl) FormatModelSetting(req *prot.SettingRequest) *models.Setting {
	if req == nil {
		return nil
	}

	values := make(models.SettingValues, 0, len(req.Values))
	for _, pair := range req.Values {
		if pair == nil {
			continue
		}
		values = append(values, models.SettingKVPair{
			Key:   pair.Key,
			Value: pair.Value,
		})
	}

	return &models.Setting{
		ID:       req.Id,
		Key:      req.Key,
		Name:     req.Name,
		Values:   values,
		IsActive: req.IsActive,
	}
}
