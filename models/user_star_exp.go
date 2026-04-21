package models

import (
	"time"
)

type UserStarExp struct {
	ID         int64     `gorm:"primaryKey;column:id" json:"id"`
	UserID     int64     `gorm:"column:user_id" json:"user_id"`
	TotalStar  int64     `gorm:"column:total_star" json:"total_star"`
	ChangeStar int64     `gorm:"column:change_star" json:"change_star"`
	TotalExp   float64   `gorm:"column:total_exp;type:numeric(15,2)" json:"total_exp"`
	ChangeExp  float64   `gorm:"column:change_exp;type:numeric(10,2)" json:"change_exp"`
	IsCurrent  bool      `gorm:"column:is_current" json:"is_current"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	Description string   `gorm:"column:description;type:text" json:"description"`
}

func (UserStarExp) TableName() string {
	return "user_star_exp"
}


