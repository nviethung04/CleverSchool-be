package requests

type UserActivatedRequest struct {
    Limit  int `form:"limit"`
    Page   int `form:"page"`
    Month  int `form:"month"`
    Year   int `form:"year"`
    RoleId int `form:"role_id"`
}
