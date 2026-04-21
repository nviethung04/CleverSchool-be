package models

func (UserAddress) TableName() string {
	return "user_address"
}

type UserAddress struct {
	ID           int    `json:"id"`
	UserId       int    `json:"user_id"`
	ProvinceCode string `json:"province_code"`
	WardCode     string `json:"ward_code"`
	Address      string `i18nKey:"address" locale:"vi"`
	ProvinceName string `i18nKey:"province_name" locale:"vi"`
	WardName     string `i18nKey:"ward_name" locale:"vi"`
	AddressEn      string `i18nKey:"address" locale:"en"`
	ProvinceNameEn string `i18nKey:"province_name" locale:"en"`
	WardNameEn     string `i18nKey:"ward_name" locale:"en"`
}

type UserAddressTranslation struct {
	Address     string
	ProvinceName string
	WardName string
}
