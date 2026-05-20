package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
)

type SettingResource interface {
	FormatSetting(setting *models.Setting) *prot.Setting
	FormatSettings(settings []*models.Setting) []*prot.Setting
	FormatModelSetting(setting *prot.SettingRequest) *models.Setting
}

type SettingResourceImpl struct{}

func NewSettingResource() SettingResource {
	return &SettingResourceImpl{}
}

func (r *SettingResourceImpl) FormatSetting(setting *models.Setting) *prot.Setting {
	if setting == nil {
		return nil
	}

	protoValues := make([]*prot.SettingValue, 0, len(setting.Values))
	for _, v := range setting.Values {
		protoValues = append(protoValues, &prot.SettingValue{
			Key:   v.Key,
			Value: v.Value,
		})
	}

	return &prot.Setting{
		Id:         setting.ID,
		Key:        setting.Key,
		Name:       setting.Name,
		Values:     protoValues,
		IsActive:   setting.IsActive,
		IsInternal: setting.IsInternal,
		CreatedAt:  setting.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  setting.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SettingResourceImpl) FormatSettings(settings []*models.Setting) []*prot.Setting {
	result := make([]*prot.Setting, 0, len(settings))
	for _, s := range settings {
		if formatted := r.FormatSetting(s); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *SettingResourceImpl) FormatModelSetting(setting *prot.SettingRequest) *models.Setting {
	if setting == nil {
		return nil
	}

	var values models.SettingValues
	for _, v := range setting.Values {
		values = append(values, models.SettingValue{
			Key:   v.Key,
			Value: v.Value,
		})
	}

	return &models.Setting{
		ID:         setting.Id,
		Key:        setting.Key,
		Name:       setting.Name,
		Values:     values,
		IsActive:   setting.IsActive,
		IsInternal: setting.IsInternal,
	}
}

