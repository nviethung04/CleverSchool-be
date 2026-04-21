package requests

type GetDistrictsRequest struct {
	ProvinceCode string `form:"province_code"`
	Limit        int    `form:"limit"`
	Page         int    `form:"page"`
}
