package models

type UserRefRole struct {
	ID         int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId   int64 `gorm:"not null" json:"user_id"`
	RoleId int64 `gorm:"not null" json:"role_id"`
}
