package models

type UserPosition struct {
	ID                 int64  `json:"id"`
	UserId             int64  `json:"user_id"`
	EmployeePositionId int64  `json:"employee_position_id"`
	Note               string `json:"note"`

	User     User             `gorm:"foreignKey:UserId"`
	Position EmployeePosition `gorm:"foreignKey:EmployeePositionId"`
}
