package dto

import "time"

type UserWithRole struct {
	ID          int       `json:"id"`
	RoleID      int       `json:"role_id"`
	SchoolID    int       `json:"school_id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Name        string    `json:"name"`
	Token       string    `json:"token"`
	ExpiresTime time.Time `json:"expires_time"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Address     string    `json:"address"`
	Status      bool      `json:"status"`
	ParentID    int       `json:"parent_id"`

	RoleName string `json:"role_name"`
}

type UserResponseLogin struct {
	Token    string `json:"token"`
	RoleId   int    `json:"role_id"`
	RoleName string `json:"role_name"`
}
