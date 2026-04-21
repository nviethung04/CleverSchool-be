package requests

type GetSchoolRequest struct {
	Limit        int    `form:"limit"`
	Page         int    `form:"page"`
	Name         string `form:"name"` // filter theo tên
	Type         string `form:"type"`
	WardCode     string `form:"ward_code"`
	ProvinceCode string `form:"province_code"`
	Status       string `form:"status"`
}
