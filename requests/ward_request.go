package requests

type GetWardsRequest struct {
	ProvinceCode string `form:"province_code"`
	Limit        int    `form:"limit"`
	Page         int    `form:"page"`
}
