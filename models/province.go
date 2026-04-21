package models

type Province struct {
	Code                   string `gorm:"primaryKey;size:20"`
	Name                   string `i18nKey:"name" locale:"vi"`
	NameEn                 string `i18nKey:"name" locale:"en"`
	FullName               string `i18nKey:"full_name" locale:"vi"`
	FullNameEn             string `i18nKey:"full_name" locale:"en"`
	CodeName               string
	AdministrativeUnitID   int32
	AdministrativeRegionID int32
}

type ProvinceTranslation struct {
	Name     string
	FullName string
}
