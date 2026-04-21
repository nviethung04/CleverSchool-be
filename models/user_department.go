package models

type UserDepartment struct {
	ID           int64  `json:"id"`
	UserId       int64  `json:"user_id"`
	DepartmentId int64  `json:"department_id"`
	Note         string `json:"note"`

	User       User       `gorm:"foreignKey:UserId"`
	Department Department `gorm:"foreignKey:DepartmentId"`
}
