package requests

type GetProvincesRequest struct {
	Limit int `form:"limit"`
	Page  int `form:"page"`
}

type GetProvinceRequest struct {
	Code string `uri:"code" binding:"required"`
}
